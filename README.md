# gnutdata-api
Provides query and retrieval REST services for USDA "Food Data Central" datasets.  Also included are  utilities for loading USDA csv files into a datastore of your choice.  Couchbase is the default datastore but it's possible without a great deal of effort to implement a MongoDb, ElasticSearch or even a relational database by implementing the ds/DataSource interface.  The steps below outline how to go about building and running the applications using Couchbase.  Additional endpoint documentation is provided by a swagger.yaml and a compiled apiDoc.html in the api/dist path.

The build requires go verion 12.  If you are using couchbase then version 6 or greater is preferred.  Both the community edition or licensed edition will work.

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

For the webserver:   
```
go build -o gnutdataserver api/main.go api/routes.go
```
and for the data loader utility:   
```
go build -o dataloader admin/loader/loader.go
```
You're free to choose different names for -o binaries if you like.  


### Step 3: Install [Couchbase](https://www.couchbase.com)     
If you do not already have access to a CouchBase instance then you will need to install at least version 6 or greater of the Community edition.     

### Step 4:  Load the USDA csv data
1. From your Couchbase console or REST API, create a bucket, e.g. gnutdata and a user, e.g. gnutadmin with the Application Access role and indexes.    Sample Couchbase API scripts are also provided in the couchbase path for these steps as well.
2. Configure config.yml (see below) for host, bucket and user id/pw values you have selected.  A template is provided to get you started.
3. Download and unzip the supporting data, BFPD, FNDDS and/or SR csv files into a location of your choice.   
4. Load the data files
```
$GOBIN/dataloader -c /path/to/config.yml -i /path/to/NUT/ -t NUT 
```
```
$GOBIN/loader -c /path/to/config.yml -i /path/to/DERV/ -t DERV
```
```
$GOBIN/loader -c /path/to/config.yml -i /path/to/FoodData_Central_branded_food_csv/ -t BFPD    
```
```
$GOBIN/loader -c /path/to/config.yml -i /path/to/FoodData_Central_survey_food_csv/ -t FNDDS  
```    
```
$GOBIN/loader -c /path/to/config.yml -i /path/to/FoodData_Central_sr_csv_2019-04-02/ -t SR
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

All data is stored in [Couchbase](http://www.couchbase.com) out of the boxfdc.  

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
$GOBIN/fdcd -d -c /path/to/config.yml -r context   
where    
  -d output debugging messages     
  -c configuration file to use (defaults to ./config.yml )      
  -p TCP port to run server (defaults to 8000)    
  -r root deployment context (v1)    
  -l send stdout/stderr to named file (defaults to /tmp/bfpd.out
 ```
## Usage    
A swagger.yaml document which fully describes the API is included in the dist path.     

### Fetch a single food  by Food Data Center id (fdcid): 
```
curl -X GET http://localhost:8000/v1/food/389714 
```
##### returns meta data only for a food   
```
curl -X GET http://localhost:8000/v1/food/389714?format=meta    
```
##### returns servings data only for a food     
```
curl -X GET http://localhost:8000/v1/food/389714?format=servings     
```   
#### returns nutrient data only for a food   
```
curl -X GET http://localhost:8000/v1/food/389714?format=nutrients   
```
### Browse foods:   
```
curl -X GET http://localhost:8000/v1/browse?page=1&max=50?format=meta&sort=foodDescription
curl -X GET http://localhost:8000/v1/browse?page=1&max=50?format=full&sort=company      
curl -X GET http://localhost:8000/v1/browse?page=1&max=50?format=nutrients    
curl -X GET http://localhost:8000/v1/browse?page=1&max=50?format=servings     
```
or      
```
http GET localhost:8000/v1/browse max=50 page=1     
```

### Search foods (GET): 
Perform a simple keyword search of the index.  Include quotes to search phrases, e.g. ?q='"bubbies homemade"'.  Limit a search to a particular field with the 'f' parameter which can be one of 'foodDescription', 'company', 'upc', or 'ingredients'.   
```
curl -X GET http://localhost:8000/v1/search?q=bread&page=1&max=100    
curl -X GET http://localhost:8000/v1/search?q=bread&f=foodDescription&page=1&max=100   
```
or    
```
http GET localhost:8000/v1/search q=bread max=50 page=1 format=servings    
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