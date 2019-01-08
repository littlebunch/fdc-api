# gnutdata-api
Provides query and retrieval REST services for USDA "Food Data Central" datasets.
## Installation
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
### Step 3:Install gnut-api services into your $GOBIN:
```
cd bfpd; go build -o $GOBIN/bfpd main.go routes.go
```
### Step 4: Install [Couchbase](https://www.couchbase.com)

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

### Start couchbase

### Start the web server:
```
$GOBIN/gnutapi -d -i -c /path/to/config.yml -r context   
where
  -d output debugging messages  
  -i initialize an authentication store
  -c configuration file to use (defaults to ./config.yml )  
  -p TCP port to run server (defaults to 8000)
  -r root deployment context (defaults to bfpd)
  -l send stdout/stderr to named file (defaults to /tmp/bfpd.out
 ```
## Usage
Authentication is required for POST, PUT and DELETE.  Use the login handler to obtain a token which then must be sent in the Authorization header as shown in the examples below.  

### Authenticate and obtain JWT token:
```
curl -X POST -H "Content-type:application/json" -d '{"password":"your-password","username":"your-user-name"}' http://localhost:8000/v1/login
```
or if you prefer [httpie](https://github.com/jakubroztocil/httpie):
```
http -v --json POST localhost:8000/v1/login username=your-password password=your-user-name
```
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
### Browse foods:
```
curl -X GET http://localhost:8000/v1/foods?page=1&max=50?format=brief
curl -X GET http://localhost:8000/v1/foods?page=1&max=50?format=full
```
or
```
http GET localhost:8000/v1/food/ max=50 page=1
```

### Perform a full text search of the index.  Include quotes to search phrases, e.g. ?q='foodDescription:"bubbies homemade"'.
```
curl -X GET http://localhost:8000/v1/foods?q=bread&page=1&max=100
```
or
```
http GET localhost:8000/v1/food q=break max=50 page=1
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
