package out

import (
	"fmt"

	"github.com/tylerrasor/defectdojo-resource/internal/concourse"
	"github.com/tylerrasor/defectdojo-resource/pkg/defectdojo_client"
)

func Put(w *concourse.Worker) error {
	request, err := DecodeToPutRequest(w)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	if request.Source.Debug {
		w.EnableDebugLog()
	}

	c := defectdojo_client.NewDefectdojoClient(request.Source.DefectDojoUrl, request.Source.ApiKey)

	w.LogDebug("looking for product profile")
	var p *defectdojo_client.Product
	p, err = c.GetProduct(request.Source.ProductName)
	if err != nil {
		if !request.Source.CreateProductIfNotExists {
			w.LogDebug("get product failed but `create_product_if_not_exist` not set")
			return fmt.Errorf("error getting product: %s", err)
		} else {
			w.LogDebug("trying to create product `%s` for product_type `%s`", request.Source.ProductName, request.Source.ProductType)
			p, err = c.CreateProduct(request.Source.ProductName, request.Source.ProductType)
			if err != nil {
				return fmt.Errorf("error creating product '%s' for product_type '%s': %s", request.Source.ProductName, request.Source.ProductType, err)
			}
		}
	}

	w.LogDebug("creating new cicd engagement")
	engagement, err := c.CreateEngagement(p, request.Params.ReportType, request.Params.CloseEngagement)
	if err != nil {
		return fmt.Errorf("error getting or creating engagement: %s", err)
	}
	w.LogDebug("built new engagement, with id: %d", engagement.EngagementId)

	workdir := w.GetWorkDir()
	full_path := fmt.Sprintf("%s/%s", workdir, request.Params.ReportPath)
	w.LogDebug("trying to read file: %s", full_path)
	bytez, err := w.ReadFile(full_path)
	if err != nil {
		return fmt.Errorf("error reading report file: %s", err)
	}

	w.LogDebug("uploading report")
	e, err := c.UploadReport(engagement.EngagementId, request.Params.ReportType, bytez)
	if err != nil {
		return fmt.Errorf("error uploading report: %s", err)
	}

	r := concourse.Response{
		Version: concourse.Version{
			EngagementId: fmt.Sprint(e.EngagementIdFromUpload),
		},
	}
	if err := w.OutputResponseToConcourse(r); err != nil {
		return fmt.Errorf("error encoding response to JSON: %s", err)
	}

	return nil
}
