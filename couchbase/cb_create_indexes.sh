#!/bin/sh
##############################################################
# Creates three indexes needed for efficient results sorting #
# Change the -u values and URL as needed.                    #
##############################################################
curl -u Administrator:password -XPOST -d 'statement=CREATE PRIMARY INDEX `idx-primary` ON `gnutbfpd` USING GSI WITH {"defer_build":true};' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fd` ON `gnutbfpd`(foodDescription) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_company` ON `gnutbfpd`(company) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:password -XPOST -d 'statement=CREATE INDEX `idx_fdcId` ON `gnutbfpd`(fdcId) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
