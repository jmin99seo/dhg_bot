package colly

import (
	"context"
	"testing"

	"github.com/gocolly/colly/v2"
)

func TestClient_SearchInvenIncidents(t *testing.T) {
	type fields struct {
		collector *colly.Collector
	}
	type args struct {
		ctx     context.Context
		keyword string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				collector: colly.NewCollector(),
			},
			args: args{
				ctx:     context.Background(),
				keyword: "숙코",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				collector: tt.fields.collector,
			}
			res, err := c.SearchInvenIncidents(tt.args.ctx, tt.args.keyword)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.SearchInvenIncidents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// t.Logf("%+v", res[0])
			for _, r := range res {
				t.Logf("%+v", r)
			}
			// if got != tt.want {
			// 	t.Errorf("Client.SearchInvenIncidents() = %v, want %v", got, tt.want)
			// }
		})
	}
}
