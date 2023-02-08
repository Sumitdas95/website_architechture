# Infrastructure

## Deploying infrastructure

Out of the box, `test-sonarqube` will deploy to the Global shard. Below will describe documentation and links to
deploy your infrastructure and project for the first time. If you require deploying to additional shards, please follow
up by reading the [Geosharding](#geosharding) below.

1. Update any references to `test-sonarqube` with the infrastructure directory to the name of your newly created
   repository. Failure to update something will result in the deployment failing due to the infrastructure already existing
   (e.g. test-sonarqube.)

2. [Geopoiesis](https://github.com/deliveroo/geopoiesis) is the repository which will
   deploy your changes. You will need to configure this repository by following
   [Step 1](https://deliveroo.atlassian.net/wiki/spaces/RLE/pages/3790209236/Terraform+-+Extracting+an+App+to+its+own+Scope#Step-1---create-the-scope-configuration-in-geopoiesis-itself)
   of the following [Guide](https://deliveroo.atlassian.net/wiki/spaces/RLE/pages/3790209236/Terraform+-+Extracting+an+App+to+its+own+Scope#Step-1---create-the-scope-configuration-in-geopoiesis-itself).
   You can find an example [pull request here](https://github.com/deliveroo/geopoiesis/pull/1136/files), merging the PR is
   not the only step.

3. Once [Geopoiesis](https://github.com/deliveroo/geopoiesis) has been deployed, you will now be able to access your
   infrastructure within `https://infrastructure.deliveroo.net`.  You can now update the related infrastructure links within
   the `PULL_REQUEST_TEMPLATE` within the `.github` directory.

4. Finally, navigate to your staging and production infrastructure links and trigger a new rollout. This will create
   your service in AWS, [Hopper](http://go/hopper) and [Supported](http://go/supported).
   Ensure to always check the release plan before accepting, and ask for support if required.

**Reminder**: The created services resources are not defined in Hopper or Geopoiesis. Its configured in your local 
hopper configuration file (`.hopper/config.yml`). Ensure to update this to reflect your requirements.

## Geosharding

You can shard your services before you shard your data. This can be beneficial by reducing your active blast radius. To
find out more, please read the existing documentation within the [Wiki](http://go/wiki),
[sharding your infrastructure](https://deliveroo.atlassian.net/wiki/spaces/GEO/pages/3755671574/Sharding+your+Infrastructure).
