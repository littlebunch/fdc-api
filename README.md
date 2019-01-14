# gnutdata-api
Provides query and retrieval REST services for USDA "Branded Food Products" datasets.  Also included are standalone admin utilities for loading USDA csv files into CouchDB.
### Step 1: Set up go environment if necessary  
Clone this repo into your [go workspace](https://golang.org/doc/code.html), e.g. $GOPATH/src/github.com/littlebunch
### Step 2: Install supporting packages as needed.  Normally, your editor should install these automatically.  The list includes:      
*[gin framework](https://github.com/gin-gonic/gin) go get github.com/gin-gonic/gin  and go get gopkg.in/appleboy/gin-jwt.v2  
*[gin-jwt](https://github.com/appleboy/gin-jwt) go get github.com/appleboy/gin-jwt       
*[bcrypt](https://godoc.org/golang.org/x/crypto/bcrypt) go get golang.org/x/crypto/bcrypt 
*[gocb]("gopkg.in/couchbase/gocb.v1") CouchBase SDK    
*[yaml](http://gopkg.in/yaml.v2) go get gopkg.in/yaml.v2       
*[endless](https://github.com/fvbock/endless) go get github.com/fvbock/endless     
*[simplejson](https://github.com/bitly/go-simplejson) go get github.com/bitly/go-simplejson 
### Step 3:Install the gnut-bfpd-api webserver and standalone loader into your $GOBIN:
```
cd $GOPATH/src/github.com/littlebunch.com/gnut-bfpd-api/api; go build -o $GOBIN/bfpd main.go
cd $GOPATH/src/github.com/littlebunch.com/gnut-bfpd-api//ingest go build -o $GOBIN/ingest ingest.go intestbfpd.go
```
### Step 4: Install [Couchbase](https://www.couchbase.com)     
If you do already have access to a CouchBase instance then you will need to download and install the Community edition.   
### Step 5:  Load the BFPD csv data
1. From your Couchbase console, create a bucket, e.g. gnutdata and a user, e.g. gnutadmin with the following roles on the bucket:  Views Reader, Query Select, Search Reader, Data Reader, Application Access.    
2. Configure config.yml (see below) for host, bucket and user id/pw values you have selected.
3. Download and unzip the BFPD csv file into a location of your choice.   
4. Run the loader:   
```
$GOBIN/ingest -c /path/to/config.yml -i /path/to/BFFD.csv -t BFPD
```
The loader can take up to an hour or two for a complete load.
5. Start the web server (see below)   

## Configuration
Configuration is minimal and can be in a YAML file or envirnoment variables which override the config file.  

YAML    
```
couchdb:
  url:  localhost
  bucket: gnutdata   //default  
  user: <your_user>
  pwd: <your_password>

```

All data is stored in [Couchbase](http://www.couchbase.com)

Environment   
```
COUCHBASE_URL=localhost
COUCHBASE_COLLECTION=bfpd
COUCHBASE_DB=foods
COUCHBASE_USER=bfpduser
COUCHBASE_PWD=bfpduser_password

```
## Running    

The instructions below assume you are deploying on a local workstation.   


### Start the web server:
```
$GOBIN/bfpd -d -c /path/to/config.yml -r context   
where
  -d output debugging messages 
  -c configuration file to use (defaults to ./config.yml )  
  -p TCP port to run server (defaults to 8000)
  -r root deployment context (defaults to bfpd)
  -l send stdout/stderr to named file (defaults to /tmp/bfpd.out
 ```
## Usage

### Fetch a single food by id: 
```
curl -X GET http://localhost:8000/v1/food/45001535  
```
#### returns meta data only for a food  
```
curl -X GET http://localhost:8000/v1/food/meta/45001535
```
#### returns servings data only for a food  
```
curl -X GET http://localhost:8000/v1/food/servings/45001535
```   
### returns nutrient data only for a food
```
curl -X GET http://localhost:8000/v1/food/nutrients/45001535
```
### Browse foods:
```
curl -X GET http://localhost:8000/v1/foods?page=1&max=50?format=meta&sort=foodDescription
curl -X GET http://localhost:8000/v1/foods?page=1&max=50?format=full&sort=company
curl -X GET http://localhost:8000/v1/foods?page=1&max=50?format=nutrients
curl -X GET http://localhost:8000/v1/foods?page=1&max=50?format=servings
```
or
```
http GET localhost:8000/v1/food/ max=50 page=1
```

### Perform a full text search of the index.  Include quotes to search phrases, e.g. ?q='foodDescription:"bubbies homemade"'.
```
curl -X GET http://localhost:8000/v1/foods/search?q=bread&page=1&max=100
```
or
```
http GET localhost:8000/v1/foods/search q=break max=50 page=1
```
### GET a list of nutrients
```
curl -X GET http://localhost:8000/v1/nutrient/list   
curl -X GET http://localhost:8000/v1/nutrient/list?nutrientno=301
curl -X GET http://localhost:8000/v1/nutrient/list?nutrient=calcium
```
or
```
http GET localhost:8000/v1/nutrient/list   
http GET localhost:8000/v1/nutrient/list?nutrientno=301
http GET localhost:8000/v1/nutrient/list?nutrient=calcium
```
