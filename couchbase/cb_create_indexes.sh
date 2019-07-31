#!/bin/sh
##############################################################
# Creates three indexes needed for efficient results sorting #
# Change the -u values and URL as needed.                    #
#                                                            #
# You will also need to execute a build when you're ready    #
# to build the indexes, i.e. after the data is loaded.
##############################################################
curl -u Administrator:password  -XPOST -d 'statement=CREATE PRIMARY INDEX `idx-primary` ON `gnutdata` USING GSI WITH {"defer_build":true};' http://localhost:8093/query/service
curl -u Administrator:password  -XPOST -d 'statement=CREATE INDEX idx_datasource ON gnutdata(dataSource) WHERE (type="FOOD") USING GSI WITH {"defer_build":true};' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fd` ON `gnutdata`(foodDescription) WHERE (type="FOOD") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fd_asc` ON `gnutdata`(foodDescription ASC) WHERE (type="FOOD") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fd_desc` ON `gnutdata`(foodDescription DESC) WHERE (type="FOOD") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_company` ON `gnutdata`(company) WHERE (type="FOOD") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_company_asc` ON `gnutdata`(company ASC) WHERE (type="FOOD") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_company_desc` ON `gnutdata`(company DESC) WHERE (type="FOOD") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fdcId` ON `gnutdata`(fdcId) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fdcId_asc` ON `gnutdata`(fdcId ASC) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fdcId_desc` ON `gnutdata`(fdcId DESC) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_foodgroup` ON `gnutdata`((`foodGroup`.`description`)) WHERE (type="FOOD") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_type` ON `gnutdata`(`type`)  USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_nutdata_query_desc` ON `gnutdata`(valuePer100UnitServing DESC,nutrientNumber DESC,Datasource DESC,unit DESC,fdcId DESC) WHERE (type = "NUTDATA") USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
# Run the build after the indexes have been created and you've loaded the data by copying and pasteing this curl call
# curl -u Administrator:password -XPOST -d 'statement=BUILD INDEX ON `gnutdata`(`idx-primary`,`idx_datasource`,`idx_fd`,`idx_fd_asc`,`idx_fd_desc`,`idx_company`,`idx_company_asc`,`idx_company_desc`,`idx_fdcId`,`idx_fdcId_asc`,`idx_fdcId_desc`,`idx_foodgroup`,`idx_type`,`idx_nutdata_query_desc`) USING GSI' http://localhost:8093/query/service