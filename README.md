# artibot

A lambda function to prune Artifactory of unused images

This function uses [github.com/lusis/go-artifactory](github.com/lusis/go-artifactory) to make an AQL search for images that have not been downloaded for a specified period. It then saves the results to a S3 bucket and makes API calls to delete those images.

These envars need to be defined in AWS Lambda:
```
ARTIFACTORY_URL      = Artifactory API URL
ARTIFACTORY_USERNAME = Artifactory username
ARTIFACTORY_PASSWORD = Artifactory password
dry_run              = True/False
repo                 = Artifactory repository name
bucket               = S3 bucket name where search results will be sent 
created              = Number of months since the image was created
downloaded           = Number of months since the image was last downloaded
modified             = Number of months since the image was last modified
```
### to do:

- add goroutine for concurrent API calls
- add support for Artifactory API tokens
- specify AWS region
