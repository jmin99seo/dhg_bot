package colly

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type InvenIncidentResult struct {
	PostURL      string `json:"post_url"`
	Title        string `json:"title"`
	Server       string `json:"server"`
	Author       string `json:"author"`
	ViewCount    int    `json:"view_count"`
	LikeCount    int    `json:"like_count"`
	DateStr      string `json:"date_str"`
	CommentCount int    `json:"comment_count"`
	HasImage     bool   `json:"has_image"`
}

func removeServerFromTitle(title, server string) string {
	t := strings.TrimLeftFunc(title, func(r rune) bool {
		return r == ' ' || r == '\n'
	})
	t = strings.TrimLeft(t, fmt.Sprintf("[%s]", server))
	t = strings.TrimLeftFunc(t, func(r rune) bool {
		return r == ' ' || r == '\n'
	})
	t = strings.TrimSpace(t)
	return t
}

func (c *client) SearchInvenIncidents(ctx context.Context, keyword string) ([]*InvenIncidentResult, error) {
	url := fmt.Sprintf("https://www.inven.co.kr/board/lostark/5355?query=list&sterm=&name=subjcont&keyword=%s", url.PathEscape(keyword))

	var results []*InvenIncidentResult

	collector := c.StartCollector()

	collector.OnHTML("#new-board > form > div > table > tbody", func(e *colly.HTMLElement) {
		children := e.DOM.ChildrenFiltered(":not(.notice)")
		children.Each(func(i int, s *goquery.Selection) {
			postURL, found := s.Find("td.tit a.subject-link").Attr("href")

			titleAnchor := s.Find("td.tit a.subject-link")

			server := titleAnchor.Find(".category").Text()
			server = strings.Trim(server, "[]")

			title := removeServerFromTitle(titleAnchor.Text(), server)

			author := s.Find("td.user span.layerNickName").Text()

			var nViewCount int
			viewCount := s.Find("td.view").Text()
			viewCount = strings.ReplaceAll(viewCount, ",", "")
			nViewCount, _ = strconv.Atoi(viewCount)

			likes := s.Find("td.reco").Text()
			nLikes, _ := strconv.Atoi(likes)

			dateStr := s.Find("td.date").Text()

			comments := s.Find("td.tit span.con-comment").Text()
			var nComments int
			if comments != "&ensp;" {
				cutStr := strings.Trim(comments, "[]")
				nComments, _ = strconv.Atoi(cutStr)
			}

			images := s.Find("td.tit").Has("span.board-img")
			var hasImage bool
			if images != nil && len(images.Nodes) > 0 {
				hasImage = true
			}

			if found {
				results = append(results, &InvenIncidentResult{
					PostURL:      postURL,
					Title:        title,
					Server:       server,
					Author:       author,
					ViewCount:    nViewCount,
					LikeCount:    nLikes,
					DateStr:      dateStr,
					CommentCount: nComments,
					HasImage:     hasImage,
				})
			}
		})
	})

	collector.Visit(url)

	collector.Wait()

	return results, nil
}

const (
	BoardNameSasagae = "로스트아크 인벤 서버 사건/사고 게시판"
)

func (c *client) SearchInvenArticles(ctx context.Context, keyword string) ([]*InvenIncidentResult, error) {
	var (
		maxResults = 50
		maxPages   = 10
	)

	collOpts := []colly.CollectorOption{
		colly.MaxDepth(1),
	}
	collector := c.StartCollector(collOpts...)

	var results []*InvenIncidentResult

	collector.OnHTML("div.section_box,noboard", func(e *colly.HTMLElement) {
		// detect if children has "ul.noresult"
		if e.DOM.Find("ul.noresult").Length() > 0 {
			fmt.Println("noresult")
			return
		}

		e.DOM.Find("ul.news_list > li.item").Each(func(_ int, el *goquery.Selection) {
			if el.Find("div.item_info > a.board").Text() == BoardNameSasagae {
				// do something with the element
				// return caption
				link := el.Find("h1 > a.name").AttrOr("href", "")
				// fmt.Println(link)
				collector.Visit(link)
			}
		})
	})

	// actual article
	collector.OnHTML("div#tbArticle", func(e *colly.HTMLElement) {
		if len(results) >= maxResults {
			return
		}

		title := e.DOM.Find("div.articleSubject h1").Text()
		dateStr := e.DOM.Find("div.articleDate").Text()
		postURL := e.Request.URL.String()
		server := e.DOM.Find("div.articleHeadMenu div.articleCategory").Text()
		server = strings.TrimSpace(server)
		server = strings.Trim(server, "[]")
		authorUnicode := e.DOM.Find("div.articleWriter span").Text()
		author, err := strconv.Unquote(`"` + authorUnicode + `"`)
		if err != nil {
			author = authorUnicode
		}

		// not a node
		var viewCount int
		hit := e.DOM.Find("div.articleHit").Contents()
		viewCountStr := hit.Text()
		viewCountStr = strings.TrimSpace(viewCountStr)
		viewCountStr = strings.TrimPrefix(viewCountStr, "조회: ")
		re := regexp.MustCompile(`(\d{1,3}(,\d{3})*)(\t|\s)*`)
		match := re.FindStringSubmatch(viewCountStr)
		if len(match) > 1 {
			nStr := strings.ReplaceAll(match[1], ",", "")
			if num, err := strconv.Atoi(nStr); err == nil {
				viewCount = num
			}
		}

		likeCountStr := e.DOM.Find("div.articleHit > span#bbsRecommendNum1").Text()
		likeCountStr = strings.TrimSpace(likeCountStr)
		likeCountStr = strings.ReplaceAll(likeCountStr, ",", "")
		likeCount, _ := strconv.Atoi(likeCountStr)

		commentCountStr := e.DOM.Find("div.topmenuInfo a > span").Text()
		commentCount, err := strconv.Atoi(commentCountStr)
		if err != nil {
			commentCount = 0
		}

		results = append(results, &InvenIncidentResult{
			PostURL:      postURL,
			Title:        title,
			DateStr:      dateStr,
			Server:       server,
			Author:       author,
			ViewCount:    viewCount,
			LikeCount:    likeCount,
			CommentCount: commentCount,
			HasImage:     false,
		})
	})

	for pageNum := 1; pageNum <= maxPages; pageNum++ {
		url := fmt.Sprintf("https://www.inven.co.kr/search/lostark/article/%s/%d", url.PathEscape(keyword), pageNum)
		collector.Visit(url)
	}

	collector.Wait()

	for _, res := range results {
		if res.ViewCount > 1000 {
			fmt.Printf("%+v\n", res)
		}
	}

	return results, nil

}
