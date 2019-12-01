# artibot

An AWS Lambda function to prune Artifactory of unused images

This function uses [github.com/lusis/go-artifactory](https://github.com/lusis/go-artifactory) to run an AQL search for images that have not been downloaded for a specified period and uploads the results to a S3 bucket. It then makes the API calls to delete those images.

These envars need to be defined in AWS Lambda:
```
dry_run                                  = True/False
repo                                     = Artifactory repository name
bucket                                   = S3 bucket name where search results will be sent
region                                   = S3 bucket region 
created                                  = Number of months since the image was created
downloaded                               = Number of months since the image was last downloaded
modified                                 = Number of months since the image was last modified
ARTIFACTORY_URL                          = Artifactory API URL
ARTIFACTORY_USERNAME                     = Artifactory username
ARTIFACTORY_PASSWORD / ARTIFACTORY_TOKEN = Artifactory password or token
```

You also need to ensure that the function has the necessary IAM permissions for the specified S3 bucket.

### to do:

- add goroutine for concurrent API calls
- print operation summary