/*
      Copyright 2024 Richie Oy

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"

	"github.com/alessio/shellescape"
)

func printWhoami(ctx context.Context, client *sts.Client) error {
	resp, err := client.GetCallerIdentity(ctx,
		&sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}

	// Consider: is resp.Userid useful for anything?
	fmt.Printf("%s\n",
		*resp.Arn)
	return nil
}

func printCredsText(creds types.Credentials) {
	fmt.Printf("%s %s %s\n",
		*creds.AccessKeyId,
		*creds.SecretAccessKey,
		*creds.SessionToken)
}

func printCredsShell(creds types.Credentials) {
	fmt.Printf(`export AWS_ACCESS_KEY_ID=%s
export AWS_SECRET_ACCESS_KEY=%s
export AWS_SESSION_TOKEN=%s
`,
		shellescape.Quote(*creds.AccessKeyId),
		shellescape.Quote(*creds.SecretAccessKey),
		shellescape.Quote(*creds.SessionToken))
}

func printCredsJson(creds types.Credentials) {
	// Compatible with aws sts get-session-token
	jsonData := map[string]interface{}{
		"Credentials": map[string]string{
			"AccessKeyId":     *creds.AccessKeyId,
			"SecretAccessKey": *creds.SecretAccessKey,
			"SessionToken":    *creds.SessionToken,
			"Expiration": creds.Expiration.Format(
				time.RFC3339),
		},
	}
	jsonOut, err := json.Marshal(jsonData)
	if err != nil {
		panic("JSON marshaling failed: " + err.Error())
	}
	fmt.Println(string(jsonOut))
}

func main() {
	var duration int
	var outfmtStr, assumeRole string
	var whoamiMode bool

	flag.IntVar(&duration, "d", 1800,
		"Validity duration of session credentials, in seconds")
	flag.StringVar(&outfmtStr, "f", "sh",
		"Output format: sh, text, json (compatible with awscli)")
	flag.StringVar(&assumeRole, "r", "",
		"Assume the specified role instead of requesting session credentials")
	flag.BoolVar(&whoamiMode, "W", false,
		"Whoami mode: print current user (other options ignored)")

	flag.Parse()

	// Minimum 900 is dictated by AWS API. The maximum is an arbitary
	// value to protect from fat-fingering.
	if duration < 900 || duration > 43200 {
		log.Fatal("session duration must be between 900 and 43200 " +
			"seconds")
	}

	var outfmt func(types.Credentials)
	switch outfmtStr {
	case "text":
		outfmt = printCredsText
	case "sh":
		outfmt = printCredsShell
	case "json":
		outfmt = printCredsJson
	default:
		log.Fatal("unknown output format: ", outfmtStr)
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal("unable to load AWS SDK config: ", err.Error())
	}

	client := sts.NewFromConfig(cfg)

	// Special mode: print current identity and exit
	if whoamiMode {
		err = printWhoami(context.WithoutCancel(ctx),
			client)
		if err != nil {
			log.Fatal("unable to fetch caller identity: ",
				err.Error())
		}
		os.Exit(0)
	}

	var creds types.Credentials
	if assumeRole == "" {
		// No role ARN: session credentials
		tokenResp, err := client.GetSessionToken(ctx, &sts.GetSessionTokenInput{
			DurationSeconds: aws.Int32(int32(duration)),
		})
		if err != nil {
			log.Fatal("unable to acquire session token: ", err.Error())
		}

		creds = *tokenResp.Credentials
	} else {
		// Has role ARN: role credentials
		assumeResp, err := client.AssumeRole(ctx, &sts.AssumeRoleInput{
			RoleArn:         &assumeRole,
			RoleSessionName: aws.String("awscredget"),
		})
		if err != nil {
			log.Fatal("unable to assume role ", assumeRole, ": ", err.Error())
		}

		creds = *assumeResp.Credentials
	}

	outfmt(creds)
}
