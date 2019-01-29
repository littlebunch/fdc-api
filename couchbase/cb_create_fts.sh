#!/bin/sh
###########################################################################
# creates a full text index required for API search                       #
# run this after data is fully loaded. change -u values and url as needed # 
###########################################################################
curl -u Administrator:password -XPUT http://localhost:8094/api/index/fts_bfpd -H 'cache-control: no-cache' -H 'content-type:application/json' -d '{ 
 "name": "fts_bfpd",
 "type": "fulltext-index",
 "params": {
  "doc_config": {
   "docid_prefix_delim": ":",
   "docid_regexp": "",
   "mode": "docid_prefix",
   "type_field": "type"
  },
  "mapping": {
   "default_analyzer": "standard",
   "default_datetime_parser": "dateTimeOptional",
   "default_field": "_all",
   "default_mapping": {
    "dynamic": true,
    "enabled": false
   },
   "default_type": "_default",
   "docvalues_dynamic": true,
   "index_dynamic": true,
   "store_dynamic": false,
   "type_field": "_type",
   "types": {
    "BFPD": {
     "dynamic": false,
     "enabled": true,
     "properties": {
      "company": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "include_in_all": true,
         "include_term_vectors": true,
         "index": true,
         "name": "company",
         "type": "text"
        }
       ]
      },
      "fdcId": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "include_in_all": true,
         "include_term_vectors": true,
         "index": true,
         "name": "fdcId",
         "type": "text"
        }
       ]
      },
      "foodDescription": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "include_in_all": true,
         "include_term_vectors": true,
         "index": true,
         "name": "foodDescription",
         "store": true,
         "type": "text"
        }
       ]
      },
      "ingredients": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "include_in_all": true,
         "include_term_vectors": true,
         "index": true,
         "name": "ingredients",
         "type": "text"
        }
       ]
      },
      "ups": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "include_in_all": true,
         "include_term_vectors": true,
         "index": true,
         "name": "ups",
         "type": "text"
        }
       ]
      }
     }
    }
   }
  },
  "store": {
   "indexType": "scorch",
   "kvStoreName": ""
  }
 },
 "sourceType": "couchbase",
 "sourceName": "gnutbfpd",
 "sourceParams": {},
 "planParams": {
  "maxPartitionsPerPIndex": 171,
  "numReplicas": 0
 }
}'
