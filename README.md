# gnutdata-api
Provides query and retrieval REST services for USDA "FoodData Central" datasets.  Also included are utilities for loading USDA csv files into a Couchbase datastore.  It's also possible without a great deal of effort to implement a MongoDb, ElasticSearch or even a relational datastore by implementing the ds/DataSource interface for your preferred platform.   

The steps below outline how to go about building and running the applications using Couchbase.  Additional endpoint documentation is provided by a swagger.yaml and a compiled apiDoc.html in the api/dist path.

The build requires go version 12.  If you are using Couchbase, then version 6 or greater is preferred.  Both the community edition or licensed edition will work.

### Step 1: Clone this repo
Clone this repo into any location other than your $GOPATH:
```
git clone git@github.com:littlebunch/gnutdata-api.git
```
and cd to the repo root, e.g.:
```
cd ~/gnutdata-api
```
      
### Step 2: Build the binaries 

The repo contains go.mod and supporting files so a build will automatically install and version all needed libraries.  If you don't want to use go mod then rm go.mod and go.sum and have at it the old-fashioned way.  For the webserver:   
```
go build -o $GOBIN/gnutdataserver api/main.go api/routes.go
```
and for the data loader utility:   
```
go build -o $GOBIN/dataloader admin/loader/loader.go
```
You're free to choose different names for -o binaries as you like.  


### Step 3: Install [Couchbase](https://www.couchbase.com)     
If you do not already have access to a CouchBase instance then you will need to install at least version 6 or greater of the Community edition.     

### Step 4:  Load the USDA csv data
1. From your Couchbase console or REST API, create a bucket, e.g. gnutdata and a user, e.g. gnutadmin with the Application Access role and indexes.    Sample Couchbase API scripts are also provided in the couchbase path for these steps as well.
2. Configure config.yml (see below) for host, bucket and user id/pw values you have selected.  A template is provided to get you started.
3. Download from https://fdc.nal.usda.gov/download-datasets.html and unzip the supporting data, BFPD, FNDDS and SR csv files into a location of your choice.   
4. Load the data files
```
$GOBIN/dataloader -c /path/to/config.yml -i /path/to/FoodData_Central_Supporting_Data_csv/ -t NUT 
```
```
$GOBIN/dataloader -c /path/to/config.yml -i /path/to/FoodData_Central_Supporting_Data_csv/ -t DERV
```
```
$GOBIN/dataloader -c /path/to/config.yml -i /path/to/FoodData_Central_branded_food_csv/ -t BFPD    
```
```
$GOBIN/dataloader -c /path/to/config.yml -i /path/to/FoodData_Central_survey_food_csv/ -t FNDDS  
```    
```
$GOBIN/dataloader -c /path/to/config.yml -i /path/to/FoodData_Central_sr_csv_2019-04-02/ -t SR
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
$GOBIN/gnutdataserver -d -c /path/to/config.yml -r context   
where    
  -d output debugging messages     
  -c configuration file to use (defaults to ./config.yml )      
  -p TCP port to run server (defaults to 8000)    
  -r root deployment context (v1)    
  -l send stdout/stderr to named file (defaults to /tmp/bfpd.out
 ```
## Usage    
A swagger.yaml document which fully describes the API is included in the dist path.     

### Fetch a single food  by Food Data Center id (fdcid=389714): 
```
curl -X GET http://localhost:8000/v1/food/389714 
```
##### returns all nutrient data for a food   
```
curl -X GET http://localhost:8000/v1/nutrient/food/389714  
```
##### returns nutrient data for a single nutrient for a food
```
curl -X GET http://localhost:8000/v1/nutrient/food/389714?n=208 
```  
### Browse foods:   
```
curl -X GET http://localhost:8000/v1/browse?page=1&max=50?sort=foodDescription
curl -X GET http://localhost:8000/v1/browse?page=1&max=50?sort=company&order=desc    
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
curl -X GET http://localhost:8000/v1/nutrients
```
```
curl -X GET http://localhost:8000/v1/nutrients?sort=nutrientno
```
```
curl -X GET http://localhost:8000/v1/nutrients?sort=name&order=desc
```
