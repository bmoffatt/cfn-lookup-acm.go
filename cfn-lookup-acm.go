// package cfn_lookup_acm implementes an AWS Cloudformation Custom Resource Lambda
//
// To use, add the following anywhere in the lambda application
//
//	import (
//		_ "github.com/bmoffatt/cfn-lookup-acm.go" // import for side effect
//	)
//
// Then, in the SAM template.yml, configure a custom resource
//
//	Acm:
//	  Type: Custom::Acm
//	  Properties:
//	    ServiceToken: !GetAtt LookupAcm.Arn
//	    DomainName: !Sub "${SubDomain}.${ApexDomain}"
//	LookupAcm:
//	  Type: AWS::Serverless::Function
//	  Metadata:
//	    BuildMethod: go1.x
//	  Properties:
//	  CodeUri: . # or wherever 🤷
//	  Environment:
//	    Variables:
//	      LOOKUP_ACM: yes # any value is fine
//	  Policies:
//	  - Statement:
//	    - Effect: Allow
//	      Action:
//	      - acm:ListCertificates
//	      Resource: "*"
package cfn_lookup_acm

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/bmoffatt/must.go"
)

func init() {
	if os.Getenv("LOOKUP_ACM") == "" || os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		return
	}
	lambda.Start(cfn.LambdaWrap(func(ctx context.Context, event cfn.Event) (string, map[string]any, error) {
		switch event.RequestType {
		case cfn.RequestDelete:
			return "the", nil, nil
		}
		var domainName = event.ResourceProperties["DomainName"]
		if domainName == "" {
			return "the", nil, errors.New("DomainName not defined")
		}
		cfg := must.Return(config.LoadDefaultConfig(ctx))
		client := acm.NewFromConfig(cfg)
		certificates := must.Return(client.ListCertificates(ctx, &acm.ListCertificatesInput{}))
		for _, certificate := range certificates.CertificateSummaryList {
			if *certificate.DomainName == domainName {
				return "the", map[string]any{"CertificateArn": certificate.CertificateArn}, nil
			}
		}
		return "the", nil, fmt.Errorf("failed to find certificate for domain %s", domainName)
	}))
}
