package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	lib "github.com/lusis/go-artifactory/artifactory.v54"
)

// aqlStats represents artifact statistics
type aqlStats struct {
	Downloaded string `json:"downloaded,omitempty"`
}

// extendedAQLFileInfo adds aqlStats to upstream struct
type extAQLFileInfo struct {
	*lib.AQLFileInfo
	Stats []aqlStats `json:"stats,omitempty"`
}

// extendedAQLResults adds aqlStats to upstream struct
type extAQLResults struct {
	*lib.AQLResults
	ExtResults []extAQLFileInfo `json:"results"`
}

// exitErrorf handles errors and exits
func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

// search finds artifacts not modified or downloaded in n months
func search(cl *lib.Client, r string, c, m, d int) ([]byte, []extAQLFileInfo, error) {

	// construct and make AQL request
	var request lib.Request
	request.Verb = "POST"
	request.Path = "/api/search/aql"
	aqlString := fmt.Sprintf(`items.find(
			{
			"$and": [
				{"repo":"%s"},
				{"created": {"$before": "%dmo"}},
				{"modified": {"$before": "%dmo"}},
				{"stat.downloaded":{"$before": "%dmo"}}
				]
			}
			).include("updated","created_by","repo","type","size",
			"name","modified_by","path","modified","id","actual_sha1",
			"created","stat.downloaded")`, r, c, m, d)

	request.Body = bytes.NewReader([]byte(aqlString))
	request.ContentType = "text/plain"

	resp, err := cl.HTTPRequest(request)
	if err != nil {
		exitErrorf("could not query Artifactory API: ", err)
	}

	// decode response bytes into json
	var res extAQLResults
	err = json.Unmarshal(resp, &res)
	if err != nil {
		exitErrorf("could not decode Artifactory response: ", err)
	}
	list := res.ExtResults

	return resp, list, nil
}

// upload AQL search results to S3
func upload(resp []byte, b, rg, r string) error {

	// put bytes in reader
	file := bytes.NewReader(resp)

	// configure s3 client
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(rg)},
	)
	if err != nil {
		exitErrorf("could not init S3 session: ", err)
	}

	// use timestamp and repo as filename
	t := time.Now()
	tf := t.Format(time.RFC3339)
	fn := (tf) + "-" + (r)

	uploader := s3manager.NewUploader(sess)

	// upload to s3
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(b),
		Key:    aws.String(fn),
		Body:   file,
	})
	if err != nil {
		exitErrorf("could not upload %q to %q, %v: ", fn, b, err)
	}

	fmt.Printf("successfully uploaded %q to %q\n", fn, b)
	return nil
}

// delete makes API calls to remove those artifacts
func delete(cl *lib.Client, list *[]extAQLFileInfo) error {

	// range over list and make delete calls to Artifactory
	for _, d := range *list {

		//construct request
		dl := []string{
			d.Repo,
			d.Path,
			d.Name,
		}
		dlj := strings.Join(dl, "/")

		var request lib.Request
		request.Verb = "DELETE"
		request.Path = "/" + (dlj)

		// make request
		_, err := cl.HTTPRequest(request)
		if err != nil {
			exitErrorf("could not delete %q: ", dlj, err)
		}
		fmt.Println("deleted: ", dlj)
	}
	return nil
}

// handler triggers search, upload and delete
func handler() error {

	// get envars
	repo := os.Getenv("repo")
	bucket := os.Getenv("bucket")
	region := os.Getenv("region")

	dry, err := strconv.ParseBool(os.Getenv("dry_run"))
	if err != nil {
		exitErrorf("could not parse envar: ", err)
	}

	created, err := strconv.Atoi(os.Getenv("created"))
	if err != nil {
		exitErrorf("could not parse envar: ", err)
	}

	modified, err := strconv.Atoi(os.Getenv("modified"))
	if err != nil {
		exitErrorf("could not parse envar: ", err)
	}

	downloaded, err := strconv.Atoi(os.Getenv("downloaded"))
	if err != nil {
		exitErrorf("could not parse envar: ", err)
	}

	// configure Artifactory client
	client, err := lib.NewClientFromEnv()
	if err != nil {
		exitErrorf("could not init Artifactory client: ", err)
	}

	// find unused artifacts and upload the list to S3
	report, list, err := search(client, repo, created, modified, downloaded)
	if err != nil {
		exitErrorf("could not list artifacts: ", err)
	}

	err = upload(report, bucket, region, repo)
	if err != nil {
		exitErrorf("could not upload report: ", err)
	}

	// delete unused artifacts
	if !dry {
		err := delete(client, &list)
		if err != nil {
			exitErrorf("could not delete artifacts: ", err)
		}
	}
	return nil
}

func main() {

	lambda.Start(handler)

}
