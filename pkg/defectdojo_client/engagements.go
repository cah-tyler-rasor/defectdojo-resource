package defectdojo_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Engagement struct {
	EngagementId     int    `json:"id,omitempty"`
	ProductId        int    `json:"product"`
	StartDate        string `json:"target_start"`
	EndDate          string `json:"target_end"`
	EngagementType   string `json:"engagement_type"`
	EngagementStatus string `json:"status,omitempty"`
	EngagementName   string `json:"name"`
}

func (c *DefectdojoClient) GetEngagement(id string) (*Engagement, error) {
	path := fmt.Sprintf("engagements/%s", id)
	params := map[string]string{}
	resp, err := c.DoGet(path, params)
	if err != nil {
		return nil, err
	}

	var e *Engagement
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&e); err != nil {
		return nil, fmt.Errorf("error decoding response: %s", err)
	}
	return e, nil
}

func (c *DefectdojoClient) CreateEngagement(p *Product, report_type string, close_engagement bool) (*Engagement, error) {
	ts := time.Now().String()
	name := fmt.Sprintf("%s - %s", report_type, ts)
	engagement_req := Engagement{
		ProductId:        p.Id,
		StartDate:        ts,
		EndDate:          ts,
		EngagementType:   "CI/CD",
		EngagementStatus: "Completed",
		EngagementName:   name,
	}

	payload, err := c.BuildJsonRequestBytez(engagement_req)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoPost("engagements", payload, APPLICATION_JSON)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	e, err := decodeToEngagement(resp)
	if err != nil {
		return nil, err
	}

	if close_engagement {
		logrus.Debugln("closing engagement because `close_engagement` set")
		path := fmt.Sprintf("engagements/%d/close", e.EngagementId)
		if resp, err := c.DoPost(path, &bytes.Buffer{}, APPLICATION_JSON); resp.StatusCode != http.StatusOK || err != nil {
			return nil, err
		}
	}

	return e, nil
}

func (c *DefectdojoClient) UploadReport(engagement_id int, report_type string, report_bytez []byte) (*Engagement, error) {
	form := map[string]string{
		"engagement": fmt.Sprint(engagement_id),
		"scan_type":  report_type,
	}

	bytez, header, err := c.BuildMultipartFormBytez(form, report_bytez)
	resp, err := c.DoPost("import-scan", bytez, header)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	e, err := decodeToEngagement(resp)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func decodeToEngagement(resp *http.Response) (*Engagement, error) {
	var e *Engagement
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&e); err != nil {
		return nil, fmt.Errorf("error decoding response: %s", err)
	}
	return e, nil
}
