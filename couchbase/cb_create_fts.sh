#!/bin/sh
###########################################################################
# creates a full text index required for API search                       #
# run this after data is fully loaded. change -u values and url as needed # 
###########################################################################
curl -u Administrator:password -XPUT http://localhost:8094/api/index/fts_bfpd -H 'cache-control: no-cache' -H 'content-type:application/json' -d '
{
  "type": "fulltext-index",
  "name": "fts_bfpd",
  "uuid": "2e333a2a97784b90",
  "sourceType": "couchbase",
  "sourceName": "gnutdata",
  "sourceUUID": "8c4a76b934d238fa36503219e0f84618",
  "planParams": {
    "maxPartitionsPerPIndex": 171
  },
  "params": {
    "doc_config": {
      "docid_prefix_delim": "",
      "docid_regexp": "",
      "mode": "type_field",
      "type_field": "type"
    },
    "mapping": {
      "analysis": {},
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
              "dynamic": false,
              "enabled": true,
              "fields": [
                {
                  "include_in_all": true,
                  "include_term_vectors": true,
                  "index": true,
                  "name": "company",
                  "store": true,
                  "type": "text"
                }
              ]
            },
            "fdcId": {
              "dynamic": false,
              "enabled": true,
              "fields": [
                {
                  "include_in_all": true,
                  "include_term_vectors": true,
                  "index": true,
                  "name": "fdcId",
                  "store": true,
                  "type": "text"
                }
              ]
            },
            "foodDescription": {
              "dynamic": false,
              "enabled": true,
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
              "dynamic": false,
              "enabled": true,
              "fields": [
                {
                  "include_in_all": true,
                  "include_term_vectors": true,
                  "index": true,
                  "name": "ingredients",
                  "store": true,
                  "type": "text"
                }
              ]
            },
            "upc": {
              "dynamic": false,
              "enabled": true,
              "fields": [
                {
                  "include_in_all": true,
                  "include_term_vectors": true,
                  "index": true,
                  "name": "upc",
                  "store": true,
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
  "sourceParams": {}
}'
