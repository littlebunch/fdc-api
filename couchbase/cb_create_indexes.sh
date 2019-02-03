#!/bin/sh
##############################################################
# Creates three indexes needed for efficient results sorting #
# Change the -u values and URL as needed.                    #
##############################################################
curl -u Administrator:maggie -XPOST -d 'statement=CREATE PRIMARY INDEX `idx-primary` ON `gnutdata` USING GSI WITH {"defer_build":true};' http://localhost:8093/query/service
curl -u Administrator:maggie -XPOST -d 'statement=CREATE INDEX `idx_fd` ON `gnutdata`(foodDescription) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:maggie -XPOST -d 'statement=CREATE INDEX `idx_company` ON `gnutdata`(company) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
curl -u Administrator:maggie -XPOST -d 'statement=CREATE INDEX `idx_fdcId` ON `gnutdata`(fdcId) USING GSI WITH {"defer_build":true}' http://localhost:8093/query/service
