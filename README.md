# FoodDataCentral-api
Provides a REST server to query and retrieve USDA [FoodData Central](https://fdc.nal.usda.gov/data-documentation.html) datasets.  You can browse foods from different sources, perform simple searches, access nutrient data for individual foods and obtain lists of foods ordered by nutrient content.  Also included is a utility for loading the USDA csv files into a Couchbase datastore.  

A quick word about Couchbase.  I've done versions of this API in MySQL, Elasticsearch and Mongo but settled on Couchbase because of [N1QL](https://www.couchbase.com/products/n1ql) and the built-in [full text search](https://docs.couchbase.com/server/6.0/fts/full-text-intro.html) engine.  I've heard it scales pretty good as well. :) It's also possible without a great deal of effort to implement a MongoDb, ElasticSearch or relational datastore by implementing the ds/DataSource interface for your preferred platform.       

The steps below outline how to go about building and running the applications using Couchbase.  Additional endpoint documentation is provided by a swagger.yaml and a compiled apiDoc.html in the [api/dist](https://github.com/littlebunch/FoodDataCentral-api/tree/master/api/dist) path.  A docker image for the web server is also available and described below.

The build requires go version 12.  If you are using Couchbase, then version 6 or greater is preferred.  Both the community edition or licensed edition will work.

### Step 1: Clone this repo
Clone this repo into any location other than your $GOPATH:
```
git clone git@github.com:littlebunch/FoodDataCentral-api.git
```
and cd to the repo root, e.g.:
```
cd ~/FoodDataCentral-api
```
      
### Step 2: Build the binaries 

The repo contains go.mod and supporting files so a build will automatically install and version all needed libraries.  If you don't want to use go mod then rm go.mod and go.sum and have at it the old-fashioned way.  For the webserver:   
```
go build -o $GOBIN/fdcapi api/main.go api/routes.go
```
and for the data loader utility:   
```
go build -o $GOBIN/fdcloader admin/loader/loader.go
```
You're free to choose different names for -o binaries as you like.  

You can also use the [Docker](https://github.com/littlebunch/FoodDataCentral-api/blob/master/docker/Dockerfile) file to create an image for the web server.

### Step 3: Install [Couchbase](https://www.couchbase.com)     
If you do not already have access to a CouchBase instance then you will need to install at least version 6 or greater of the Community edition.  There are a number of easy deployment [options](https://resources.couchbase.com/cloud-partner-gcp/docs-deploy-gcp) from a local workstation, docker or the public cloud.  Checkout the latter from [Google](https://resources.couchbase.com/cloud-partner-gcp/docs-deploy-gcp), [Amazon](https://resources.couchbase.com/cloud-partner-gcp/docs-deploy-gcp) and [Azure](https://resources.couchbase.com/cloud-partner-gcp/docs-deploy-gcp).     

### Step 4:  Load the USDA csv data
1. From your Couchbase console or REST API, create a bucket, e.g. gnutdata and a user, e.g. gnutadmin with the Application Access role and indexes.    Sample Couchbase API scripts are also provided in the couchbase path for these steps as well.
2. Configure config.yml (see below) for host, bucket and user id/pw values you have selected.  A template is provided to get you started.
3. Download from https://fdc.nal.usda.gov/download-datasets.html and unzip the supporting data, BFPD, FNDDS and SR csv files into a location of your choice.   
4. Load the data files
```
$GOBIN/fdcloader -c /path/to/config.yml -i /path/to/FoodData_Central_Supporting_Data_csv/ -t NUT 
```
```
$GOBIN/fdcloader -c /path/to/config.yml -i /path/to/FoodData_Central_Supporting_Data_csv/ -t DERV
```
```
$GOBIN/fdcloader -c /path/to/config.yml -i /path/to/FoodData_Central_branded_food_csv/ -t BFPD    
```
```
$GOBIN/fdcloader -c /path/to/config.yml -i /path/to/FoodData_Central_survey_food_csv/ -t FNDDS  
```    
```
$GOBIN/fdcloader -c /path/to/config.yml -i /path/to/FoodData_Central_sr_csv_2019-04-02/ -t SR
``` 

5. Start the web server (see below)   

## Configuration     
Configuration is minimal and can be in a YAML file or envirnoment variables which override the config file.   

```
couchdb:   
  url:  localhost   
  bucket: gnutdata   //default  bucket    
  fts: fd_food  // default full-text index   
  user: <your_user>    
  pwd: <your_password>    

```

All data is stored in [Couchbase](http://www.couchbase.com) out of the box.  

Environment   
```
COUCHBASE_URL=localhost   
COUCHBASE_BUCKET=gnutdata   
COUCHBASE_FTSINDEX=fd_food   
COUCHBASE_USER=user_name   
COUCHBASE_PWD=user_password   
```
## Running    

The instructions below assume you are deploying on a local workstation.   


### Start the web server:    
```
$GOBIN/nutapi -d -c /path/to/config.yml -r context   
where    
  -d output debugging messages     
  -c configuration file to use (defaults to ./config.yml )      
  -p TCP port to run server (defaults to 8000)    
  -r root deployment context (v1)    
  -l send stdout/stderr to named file (defaults to /tmp/bfpd.out
 ```
 
Or, run from docker.io (you will need docker installed):
 ```
 docker run --rm -it -p 8000:8000 --env-file=./docker.env littlebunch/fdcapi
```
You will need to pass in the Couchbase configuration as environment variables described above.  The easiest way to do this is in a file of which a sample is provided in the repo's docker path.
   
## Usage    
A swagger.yaml document which fully describes the API is included in the dist path.     

### Fetch a single food  by FoodData Central id=389714: 
```
curl -X GET http://localhost:8000/v1/food/389714 
```
##### returns all nutrient data for a food   
```
curl -X GET http://localhost:8000/v1/nutrients/food/389714  
```
##### returns nutrient data for a single nutrient for a food
```
curl -X GET http://localhost:8000/v1/nutrients/food/389714?n=208 
```  
### Browse foods:   
```
curl -X GET http://localhost:8000/v1/foods/browse?page=1&max=50?sort=foodDescription
curl -X GET http://localhost:8000/v1/foods/browse?page=1&max=50?sort=company&order=desc    
```

### Search foods (GET): 
Perform a simple keyword search of the index.  Include quotes to search phrases, e.g. ?q='"bubbies homemade"'.  Limit a search to a particular field with the 'f' parameter which can be one of 'foodDescription', 'company', 'upc', or 'ingredients'.   
```
curl -X GET http://localhost:8000/v1/search?q=bread&page=1&max=100    
curl -X GET http://localhost:8000/v1/search?q=bread&f=foodDescription&page=1&max=100   
```

### Search foods (POST):
Perform a string search for 'raw broccoli' in the foodDescription field:   
```
curl -XPOST http://localhost:8000/v1/search -d '{"q":"brocolli raw","searchfield":"foodDescription","max":50,"page":0}'
```
Perform a WILDCARD search for company names that match ro*nd*:
```
curl -XPOST http://localhost:8000/v1/search -d '{"q":"ro*nd*","searchfield":"company","searchtype":"WILDCARD","max":50,"page":0}'
```
Perform a PHRASE search for an exact match on "broccoli florets" in the "ingredients field:
```
curl -XPOST http://localhost:8000/v1/search -d '{"q":"broccoli raw","searchfield":"ingredients","searchfield":"PHRASE","max":50,"page":0}'
```
### Fetch the nutrients dictionary
```
curl -X GET http://localhost:8000/v1/nutrients/browse
```
```
curl -X GET http://localhost:8000/v1/nutrients/browse?sort=nutrientno
```
```
curl -X GET http://localhost:8000/v1/nutrients/browse?sort=name&order=desc
```
### Run a nutrient report sorted in descending order by nutrient value per 100 units of measure 
Find foods which have a value for nutrient 208 (Energy KCAL) between 100 and 250 per 100 grams 
```
curl -X POST http://localhost:8000/v1/nutrients/report -d '{"nutrientno":207,"valueGTE":10,"valueLTE":50}'
```
Find Branded Food Products which have a nutrient value between 5 and 10 MG per 100 grams caffiene 
```
curl -X POST http://localhost:8000/v1/nutrients/report -d '{"nutrientno":262,"valueGTE":5,"valueLTE":10,"source":"BFPD"}'
```
