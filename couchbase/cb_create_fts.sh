#!/bin/sh
###########################################################################
# creates a full text index required for API search                       #
# run this after data is fully loaded. change -u values and url as needed # 
###########################################################################
curl -u Administrator:password -XPUT http://localhost:8094/api/index/fts_bfpd -H 'cache-control: no-cache' -H 'content-type:application/json' -d '{ 
 "name": "fts_bfpd",
 "type": "fulltext-index",
 "params": {
  "mapping": {
   "types": {
    "BFPD": {
     "enabled": true,
     "dynamic": false,
     "properties": {
      "ingredients": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "name": "ingredients",
         "type": "text",
         "store": false,
         "index": true,
         "include_term_vectors": true,
         "include_in_all": true
        }
       ]
      },
      "company": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "name": "company",
         "type": "text",
         "store": false,
         "index": true,
         "include_term_vectors": true,
         "include_in_all": true
        }
       ]
      },
      "ups": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "name": "ups",
         "type": "text",
         "store": false,
         "index": true,
         "include_term_vectors": true,
         "include_in_all": true
        }
       ]
      },
      "foodDescription": {
       "enabled": true,
       "dynamic": false,
       "fields": [
        {
         "name": "foodDescription",
         "type": "text",
         "store": false,
         "index": true,
         "include_term_vectors": true,
         "include_in_all": true
        }
       ]
      }
     }
    }
   },
   "default_mapping": {
    "enabled": false,
    "dynamic": true
   },
   "default_type": "_default",
   "default_analyzer": "standard",
   "default_datetime_parser": "dateTimeOptional",
   "default_field": "_all",
   "store_dynamic": false,
   "index_dynamic": true
  },
  "store": {
   "indexType": "scorch",
   "kvStoreName": ""
  },
  "doc_config": {
   "mode": "type_field",
   "type_field": "type",
   "docid_prefix_delim": "",
   "docid_regexp": ""
  }
 },
 "sourceType": "couchbase",
 "sourceName": "gnutbfpd",
 "sourceUUID": "31cc474496e50b5284bac5f6882b617f",
 "sourceParams": {},
 "planParams": {
  "maxPartitionsPerPIndex": 171,
  "numReplicas": 0
 },
 "uuid": ""
}'
