package colly

import (
	"context"
	"fmt"
	"net/url"
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

func (c *Client) SearchInvenIncidents(ctx context.Context, keyword string) ([]*InvenIncidentResult, error) {
	url := fmt.Sprintf("https://www.inven.co.kr/board/lostark/5355?query=list&sterm=&name=subjcont&keyword=%s", url.PathEscape(keyword))

	var results []*InvenIncidentResult

	c.collector.OnHTML("#new-board > form > div > table > tbody", func(e *colly.HTMLElement) {
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

	c.collector.Visit(url)

	c.collector.Wait()

	return results, nil
}
