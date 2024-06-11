# Changelog

## v1.7.0

BUGFIXES

* [#1394](https://github.com/zkMeLabs/mechain-storage-provider/pull/1394)  fix: pick new gvg when retry failed replicate piece task
* [#1391](https://github.com/zkMeLabs/mechain-storage-provider/pull/1391)  fix: check if it is AgentUploadTask
* [#1390](https://github.com/zkMeLabs/mechain-storage-provider/pull/1390)  fix: delegate upload param check
* [#1389](https://github.com/zkMeLabs/mechain-storage-provider/pull/1389)  fix: delegate upload param check
* [#1387](https://github.com/zkMeLabs/mechain-storage-provider/pull/1387)  fix: upgrade deps for fixing vulnerabilities
* [#1386](https://github.com/zkMeLabs/mechain-storage-provider/pull/1386)  fix: check if BucketExtraInfo is nil
* [#1384](https://github.com/zkMeLabs/mechain-storage-provider/pull/1384)  fix: fix db override

FEATURES

* [#1392](https://github.com/zkMeLabs/mechain-storage-provider/pull/1392)  feat: provide recommended vgf

## v1.6.0

BUGFIXES

* [#1375](https://github.com/zkMeLabs/mechain-storage-provider/pull/1375)  fix: fix GC issue
* [#1379](https://github.com/zkMeLabs/mechain-storage-provider/pull/1379)  perf: replicate piece
* [#1378](https://github.com/zkMeLabs/mechain-storage-provider/pull/1378)  fix: deletion on checksum not fully performed and fail to override duplicate checksum

FEATURES

* [#1354](https://github.com/zkMeLabs/mechain-storage-provider/pull/1354)  feat: simplify off-chain-auth
* [#1353](https://github.com/zkMeLabs/mechain-storage-provider/pull/1353)  feat: Primary SP as the upload agent
* [#1351](https://github.com/zkMeLabs/mechain-storage-provider/pull/1351)  feat: sp monthly free quota

## v1.5.0

BUGFIXES

* [#1344](https://github.com/zkMeLabs/mechain-storage-provider/pull/1344)  fix: exclude timestamp from recovery piece key to avoid duplicate requests
* [#1342](https://github.com/zkMeLabs/mechain-storage-provider/pull/1342)  fix: Upload_Done task loaded from DB does not have GVG info
* [#1336](https://github.com/zkMeLabs/mechain-storage-provider/pull/1336)  fix: 500 error code when bucket is deleted
* [#1334](https://github.com/zkMeLabs/mechain-storage-provider/pull/1334)  fix: gc object retry due to deleted bucket

FEATURES

* [#1337](https://github.com/zkMeLabs/mechain-storage-provider/pull/1337)  feat: atomic object update
* [#1345](https://github.com/zkMeLabs/mechain-storage-provider/pull/1345)  feat: update gvg save logic
* [#1340](https://github.com/zkMeLabs/mechain-storage-provider/pull/1340)  feat: add more info in status api

## v1.4.0

BUGFIXES

* [#1347](https://github.com/zkMeLabs/mechain-storage-provider/pull/1347)  fix: init the bucket traffic when NotifyPreMigrateBucketAndDeductQuota
* [#1343](https://github.com/zkMeLabs/mechain-storage-provider/pull/1343)  fix: correct the bucket quota formula when pre-check in bucket migration
* [#1341](https://github.com/zkMeLabs/mechain-storage-provider/pull/1341)  fix: let bucket migration be able to start without a pre-read of bucket content
* [#1333](https://github.com/zkMeLabs/mechain-storage-provider/pull/1333)  fix: Not allow to upload piece/object for a fully-uploaded object
* [#1314](https://github.com/zkMeLabs/mechain-storage-provider/pull/1314)  fix: special object download issue
* [#1315](https://github.com/zkMeLabs/mechain-storage-provider/pull/1315)  fix: blocksyncer sequential processing
* [#1316](https://github.com/zkMeLabs/mechain-storage-provider/pull/1316)  fix: fix complete swap in

FEATURES

* [#1250](https://github.com/zkMeLabs/mechain-storage-provider/pull/1250) feat: bucket migration state improvement
* [#1298](https://github.com/zkMeLabs/mechain-storage-provider/pull/1298) feat: invalid object name

## v1.3.1

This release contains 1 bugfix.

BUGFIXES

* [#1333](https://github.com/zkMeLabs/mechain-storage-provider/pull/1333) fix: Not allow to upload piece/object for a fully-uploaded object

## v1.3.0

This release contains 2 bugfix and 2 feature.

BUGFIXES

* [#1304](https://github.com/zkMeLabs/mechain-storage-provider/pull/1304) fix: fix sp unhealthy bug
* [#1311](https://github.com/zkMeLabs/mechain-storage-provider/pull/1311) fix: fix health check

FEATURES

* [#1279](https://github.com/zkMeLabs/mechain-storage-provider/pull/1279) feat: re-implementation of sp exit
* [#1312](https://github.com/zkMeLabs/mechain-storage-provider/pull/1312) perf: blocksyncer sql commit

## v1.2.5

This release contains 1 bugfixes.

BUGFIXES

* [#1305](https://github.com/zkMeLabs/mechain-storage-provider/pull/1305) fix: secondary SP might need to wait object meta due to latency

## v1.2.4

This release contains 2 bugfixes.

BUGFIXES

* [#1230](https://github.com/zkMeLabs/mechain-storage-provider/pull/1230) fix: fix gvg not exist
* [#1299](https://github.com/zkMeLabs/mechain-storage-provider/pull/1299) fix: fix: unhealthy sps clean bug

## v1.2.3

This release contains 1 bugfix and 1 feature.

BUGFIXES

* [#1289](https://github.com/zkMeLabs/mechain-storage-provider/pull/1289) fix: replicate failed sp idx not included in local task

FEATURES

* [#1291](https://github.com/zkMeLabs/mechain-storage-provider/pull/1291) feat: add more log

## v1.2.2

BUGFIXES

* [#1284](https://github.com/zkMeLabs/mechain-storage-provider/pull/1284) fix: remove tags field when creating object/group/bucket

## v1.2.1

BUGFIXES

* [#1282](https://github.com/zkMeLabs/mechain-storage-provider/pull/1282) fix: delete prefix tree slow query issue

## v1.2.0

FEATURES

* [#1263](https://github.com/zkMeLabs/mechain-storage-provider/pull/1263) feat: add Tags in object/bucket/group
* [#1260](https://github.com/zkMeLabs/mechain-storage-provider/pull/1260) feat: add task retry scheduler
* [#1258](https://github.com/zkMeLabs/mechain-storage-provider/pull/1258) add deposit and delete operation to GVG in signer and fix a few issue
* [#1190](https://github.com/zkMeLabs/mechain-storage-provider/pull/1190) gc for zombie piece & metaTask & bucket migration

BUGFIXES

* [#1257](https://github.com/zkMeLabs/mechain-storage-provider/pull/1257) fix: httpcode error
* [#1259](https://github.com/zkMeLabs/mechain-storage-provider/pull/1259) fix: refine error message when object status is unexpected
* [#1256](https://github.com/zkMeLabs/mechain-storage-provider/pull/1256) fix: fix the slow xml marshaling when returning a large list of objects
* [#1266](https://github.com/zkMeLabs/mechain-storage-provider/pull/1266) fix: queried gvg might be nil
* [#1270](https://github.com/zkMeLabs/mechain-storage-provider/pull/1270) fix: list object by bucket name bug
* [#1277](https://github.com/zkMeLabs/mechain-storage-provider/pull/1277) fix: group members api incorrect tags/source_type/extra response

## V1.1.0

FEATURES

* [#1240](https://github.com/zkMeLabs/mechain-storage-provider/pull/1240) feat: add a DataMigrationRecord table to save the process record for data migration tasks
* [#1218](https://github.com/zkMeLabs/mechain-storage-provider/pull/1218) feat: add sp health for pick sp
* [#1237](https://github.com/zkMeLabs/mechain-storage-provider/pull/1237) feat: signer module adds metrics for each rpc interface
* [#1227](https://github.com/zkMeLabs/mechain-storage-provider/pull/1227) feat: add reject bucket migration
* [#1224](https://github.com/zkMeLabs/mechain-storage-provider/pull/1224) feat: add gnfd cmd to query sq incomes
* [#1223](https://github.com/zkMeLabs/mechain-storage-provider/pull/1223) feat: update storage size for bucket apis
* [#1118](https://github.com/zkMeLabs/mechain-storage-provider/pull/1118) feat: Quota improvement for bucket migration
* [#1201](https://github.com/zkMeLabs/mechain-storage-provider/pull/1201) feat: sp adds http probe to improve service stability
* [#1207](https://github.com/zkMeLabs/mechain-storage-provider/pull/1207) feat: update recover object command for k8s job
* [#1197](https://github.com/zkMeLabs/mechain-storage-provider/pull/1197) feat: add golang runtime metrics, process metrics and std lib db metrics
* [#1167](https://github.com/zkMeLabs/mechain-storage-provider/pull/1167) feat: add error code in http metrics to help locate problems

BUGFIXES

* [#1245](https://github.com/zkMeLabs/mechain-storage-provider/pull/1245) fix: fix sp healthy checker dead lock bug
* [#1228](https://github.com/zkMeLabs/mechain-storage-provider/pull/1228) fix: primary sp should check received ssp signature before sealing
* [#1229](https://github.com/zkMeLabs/mechain-storage-provider/pull/1229) fix: sp should resume picking VGF regardless of RPC error
* [#1198](https://github.com/zkMeLabs/mechain-storage-provider/pull/1198) fix: refine the error msg for ErrInvalidExpiryDate header and parameter
* [#1222](https://github.com/zkMeLabs/mechain-storage-provider/pull/1222) fix: bucket size write DB failed
* [#1221](https://github.com/zkMeLabs/mechain-storage-provider/pull/1221) fix: gateway module remove dependency on spdb
* [#1220](https://github.com/zkMeLabs/mechain-storage-provider/pull/1220) fix: fix metrics and pprof disable bug
* [#1213](https://github.com/zkMeLabs/mechain-storage-provider/pull/1213) fix: update expiration new logic UT
* [#1215](https://github.com/zkMeLabs/mechain-storage-provider/pull/1215) fix: delete aliyunfs code
* [#1170](https://github.com/zkMeLabs/mechain-storage-provider/pull/1170) fix: skip error when object is deleted
* [#1212](https://github.com/zkMeLabs/mechain-storage-provider/pull/1212) fix: return 400 when getting invalid header for getNonceAPI
* [#1203](https://github.com/zkMeLabs/mechain-storage-provider/pull/1203) fix: error code when invalid signature
* [#1184](https://github.com/zkMeLabs/mechain-storage-provider/pull/1184) fix: fix resource leak
* [#1204](https://github.com/zkMeLabs/mechain-storage-provider/pull/1204) fix: remove the order by clause for discontinuing function
* [#1199](https://github.com/zkMeLabs/mechain-storage-provider/pull/1199) fix: fix time.Since called in defer, it should be used in defer func
* [#1183](https://github.com/zkMeLabs/mechain-storage-provider/pull/1183) fix: adjust sp command description
* [#1180](https://github.com/zkMeLabs/mechain-storage-provider/pull/1180) fix: replace record not found with empty array

## v1.0.5

This release contains 2 bugfix.

BUGFIXES

* [#1229](https://github.com/zkMeLabs/mechain-storage-provider/pull/1229) fix: sp should resume picking VGF regardless of RPC error
* [#1239](https://github.com/zkMeLabs/mechain-storage-provider/pull/1239) fix: add timeout for RPC calls to Secondary SP

## v1.0.4

This release contains 1 bugfix.

BUGFIXES

* [#1231](https://github.com/zkMeLabs/mechain-storage-provider/pull/1231) fix: fix: add config sp blacklist

## v1.0.3

This release contains 1 bugfix.

BUGFIXES

* [#1214](https://github.com/zkMeLabs/mechain-storage-provider/pull/1214) fix: abort replication task if an object sealed

## v1.0.2

This release contains 2 bugfixes and 1 feature.

FEATURES

* [#1200](https://github.com/zkMeLabs/mechain-storage-provider/pull/1200) feat: add index to update at

BUGFIXES

* [#1189](https://github.com/zkMeLabs/mechain-storage-provider/pull/1189) fix: add challenge v2 api to avoid piece hash http header too large
* [#1195](https://github.com/zkMeLabs/mechain-storage-provider/pull/1195) fix:bs job connect db bug

## v1.0.1

FEATURES

* [#1181](https://github.com/zkMeLabs/mechain-storage-provider/pull/1181) feat: api rate limiter refactor
* [#1188](https://github.com/zkMeLabs/mechain-storage-provider/pull/1188) feat:blocksyncer fix data command

BUGFIXES

* [#1186](https://github.com/zkMeLabs/mechain-storage-provider/pull/1186) fix:batch delete group member
* [#1187](https://github.com/zkMeLabs/mechain-storage-provider/pull/1187) fix:payment refundable update

## v1.0.0

This is the first official version for the main-net deployment.

## v0.2.6-hf.2

BUGFIX

* [#1171](https://github.com/zkMeLabs/mechain-storage-provider/pull/1171) fix: fix init the NotAvailableSpIdx

## v0.2.6-hf.1

BUGFIX

* [#1166](https://github.com/zkMeLabs/mechain-storage-provider/pull/1166) fix: delete group fix job
* [#1160](https://github.com/zkMeLabs/mechain-storage-provider/pull/1160) fix: fix replication failed SP err recording issue
* [#1159](https://github.com/zkMeLabs/mechain-storage-provider/pull/1159) fix: list objects by gvg sql bug
* [#1165](https://github.com/zkMeLabs/mechain-storage-provider/pull/1165) fix: verify permission 500 code and update UT
* [#1155](https://github.com/zkMeLabs/mechain-storage-provider/pull/1155) fix: fix manager duplicate entry when inserting upload progress
* [#1158](https://github.com/zkMeLabs/mechain-storage-provider/pull/1158) fix: add xml response for rate limit error
* [#1156](https://github.com/zkMeLabs/mechain-storage-provider/pull/1156) fix: recovery object status should be sealed
* [#1154](https://github.com/zkMeLabs/mechain-storage-provider/pull/1154) fix: add ut for universal endpoint

## v0.2.6

TEST

* [#1157](https://github.com/zkMeLabs/mechain-storage-provider/pull/1157) test: add metadata api UT

## v0.2.6-alpha.2

BUGFIX

* [#1151](https://github.com/zkMeLabs/mechain-storage-provider/pull/1151) fix: update new logic of metadata apis

## v0.2.6-alpha.1

FEATURES

* [#1109](https://github.com/zkMeLabs/mechain-storage-provider/pull/1109) feat: add migrate piece auth check
* [#1131](https://github.com/zkMeLabs/mechain-storage-provider/pull/1131) feat: limit creating object and migrating bucket approval

BUGFIX

* [#1130](https://github.com/zkMeLabs/mechain-storage-provider/pull/1130) fix: defer returning err until all goroutines are done
* [#1134](https://github.com/zkMeLabs/mechain-storage-provider/pull/1134) fix: readme
* [#1139](https://github.com/zkMeLabs/mechain-storage-provider/pull/1139) fix: refactoring code for getObject and Universal apis
* [#1140](https://github.com/zkMeLabs/mechain-storage-provider/pull/1140) fix: add replicate check permission
* [#1142](https://github.com/zkMeLabs/mechain-storage-provider/pull/1142) fix: resumable queue bug
* [#1145](https://github.com/zkMeLabs/mechain-storage-provider/pull/1145) fix: policy statement is empty
* [#1146](https://github.com/zkMeLabs/mechain-storage-provider/pull/1146) fix: fix integrity hash check
* [#1147](https://github.com/zkMeLabs/mechain-storage-provider/pull/1147) fix: fix uploader server read closed channel

TEST

* [#1143](https://github.com/zkMeLabs/mechain-storage-provider/pull/1143) test: add blocksyncer ut case

## v0.2.5

BUGFIX

* [#1135](https://github.com/zkMeLabs/mechain-storage-provider/pull/1135) fix: effect allow issue

## v0.2.5-alpha.3

FEATURES

* [#1129](https://github.com/zkMeLabs/mechain-storage-provider/pull/1129) feat: add back metadata go routine listener
* [#1128](https://github.com/zkMeLabs/mechain-storage-provider/pull/1128) feat: support expire time check of SP task info

BUGFIX

* [#1092](https://github.com/zkMeLabs/mechain-storage-provider/pull/1092) fix: support reimburse quota if download consumed extra quota
* [#1127](https://github.com/zkMeLabs/mechain-storage-provider/pull/1127) fix:bs groups unique index
* [#1125](https://github.com/zkMeLabs/mechain-storage-provider/pull/1125) fix: expiration time for statement

TEST

* [#1126](https://github.com/zkMeLabs/mechain-storage-provider/pull/1126) test: modular/executor pkg adds UTs

## v0.2.5-alpha.2

FEATURES

* [#1112](https://github.com/zkMeLabs/mechain-storage-provider/pull/1112) feat: support GfSpGetSPMigrateBucketNumber

BUGFIX

* [#1121](https://github.com/zkMeLabs/mechain-storage-provider/pull/1121) fix: fix tls MinVersion to 1.2
* [#1119](https://github.com/zkMeLabs/mechain-storage-provider/pull/1119) fix: stop all replicate jobs and done replicate job by context
* [#1117](https://github.com/zkMeLabs/mechain-storage-provider/pull/1117) fix: include private cause bucket not found

TEST

* [#1115](https://github.com/zkMeLabs/mechain-storage-provider/pull/1115) test: modular/gater pkg adds UTs part II

## v0.2.5-alpha.1

FEATURES

* [#1029](https://github.com/zkMeLabs/mechain-storage-provider/pull/1029) feat: use aliyun oss sdk to visit oss
* [#1042](https://github.com/zkMeLabs/mechain-storage-provider/pull/1042) feat: add go routine metrics
* [#1057](https://github.com/zkMeLabs/mechain-storage-provider/pull/1057) feat: list object policies and add number of group members
* [#1067](https://github.com/zkMeLabs/mechain-storage-provider/pull/1067) feat: Docker image distroless update
* [#1111](https://github.com/zkMeLabs/mechain-storage-provider/pull/1111) feat: only persist the init off-chain-auth record when first updating
* [#1103](https://github.com/zkMeLabs/mechain-storage-provider/pull/1103) feat: list groups by ids
* [#1088](https://github.com/zkMeLabs/mechain-storage-provider/pull/1088) feat: recover object list
* [#1077](https://github.com/zkMeLabs/mechain-storage-provider/pull/1077) feat: support ListUserPaymentAccounts & ListPaymentAccountStreams

BUGFIX

* [#1039](https://github.com/zkMeLabs/mechain-storage-provider/pull/1039) fix: remove funding key in singer module
* [#1053](https://github.com/zkMeLabs/mechain-storage-provider/pull/1053) fix: add more metrics log for PickUpTask
* [#1054](https://github.com/zkMeLabs/mechain-storage-provider/pull/1054) fix: fix quota db to support get quota by month
* [#1062](https://github.com/zkMeLabs/mechain-storage-provider/pull/1062) fix: add fingerprint to approval key
* [#1071](https://github.com/zkMeLabs/mechain-storage-provider/pull/1071) fix: cancel migrate bucket bug
* [#1074](https://github.com/zkMeLabs/mechain-storage-provider/pull/1074) fix: set correct Content-Disposition when downloading an object
* [#1104](https://github.com/zkMeLabs/mechain-storage-provider/pull/1104) fix: refactor quota  table
* [#1107](https://github.com/zkMeLabs/mechain-storage-provider/pull/1107) fix: api rate limiter path sequence
* [#1106](https://github.com/zkMeLabs/mechain-storage-provider/pull/1106) fix: sec issue about conversion alerts
* [#1089](https://github.com/zkMeLabs/mechain-storage-provider/pull/1089) fix: sec issue about conversion alerts
* [#1098](https://github.com/zkMeLabs/mechain-storage-provider/pull/1098) fix: self sp id retrieval
* [#1097](https://github.com/zkMeLabs/mechain-storage-provider/pull/1097) fix: fix group ExpirationTime bug
* [#1093](https://github.com/zkMeLabs/mechain-storage-provider/pull/1093) fix: verify permission if expiration time = 0 bug
* [#1087](https://github.com/zkMeLabs/mechain-storage-provider/pull/1087) fix: upgrade libp2p and cosmos-sdk version to solve security issues
* [#1084](https://github.com/zkMeLabs/mechain-storage-provider/pull/1084) fix: resumable upload support 64g file
* [#776](https://github.com/zkMeLabs/mechain-storage-provider/pull/776) fix: code security check
* [#1078](https://github.com/zkMeLabs/mechain-storage-provider/pull/1078) fix: add more resource manager log

TEST

* [#1032](https://github.com/zkMeLabs/mechain-storage-provider/pull/1032) test: modular/approver pkg adds UTs
* [#1035](https://github.com/zkMeLabs/mechain-storage-provider/pull/1035) test: metadata bsdb UT
* [#1040](https://github.com/zkMeLabs/mechain-storage-provider/pull/1040) test: add cmd ut
* [#1046](https://github.com/zkMeLabs/mechain-storage-provider/pull/1046) test: add downloader ut
* [#1068](https://github.com/zkMeLabs/mechain-storage-provider/pull/1068) test: gfspapp pkg adds UTs part II
* [#1073](https://github.com/zkMeLabs/mechain-storage-provider/pull/1073) test: gfspconfig and gfsppieceop pkg add UTs
* [#1108](https://github.com/zkMeLabs/mechain-storage-provider/pull/1108) test: modular/gater pkg adds UTs
* [#1095](https://github.com/zkMeLabs/mechain-storage-provider/pull/1095) test: modular/uploader pkg adds UTs
* [#1096](https://github.com/zkMeLabs/mechain-storage-provider/pull/1096) test: base/gfsprcmgr pkg adds UTs
* [#1090](https://github.com/zkMeLabs/mechain-storage-provider/pull/1090) test: base/gfspclient pkg adds UTs
* [#1082](https://github.com/zkMeLabs/mechain-storage-provider/pull/1082) test: base/types directory adds UTs
* [#1076](https://github.com/zkMeLabs/mechain-storage-provider/pull/1076) test: gfsptqueue and gfspvgmgr pkg add UTs
* [#1060](https://github.com/zkMeLabs/mechain-storage-provider/pull/1060) test: add receiver ut
* [#1086](https://github.com/zkMeLabs/mechain-storage-provider/pull/1086) test: add blocksyncer e2e test

## v0.2.4-alpha.9

FEATURES

* [#989](https://github.com/zkMeLabs/mechain-storage-provider/pull/989) feat: impl group apis and fix verify permission bug
* [#1008](https://github.com/zkMeLabs/mechain-storage-provider/pull/1008) feat: change auth api response from json to xml
* [#1010](https://github.com/zkMeLabs/mechain-storage-provider/pull/1010) feat:blocksyncer add realtime mode
* [#1015](https://github.com/zkMeLabs/mechain-storage-provider/pull/1015) feat: retrieve groups where the user is the owner
* [#1012](https://github.com/zkMeLabs/mechain-storage-provider/pull/1012) feat: error handle updates to provide useful messages
* [#1025](https://github.com/zkMeLabs/mechain-storage-provider/pull/1025) feat: change json response body to xml

BUGFIX

* [#999](https://github.com/zkMeLabs/mechain-storage-provider/pull/999) fix: fix bug for metadata crash
* [#1000](https://github.com/zkMeLabs/mechain-storage-provider/pull/1000) fix: fix occasional compile error
* [#1006](https://github.com/zkMeLabs/mechain-storage-provider/pull/1006) fix: rename api name and replace post to get
* [#1014](https://github.com/zkMeLabs/mechain-storage-provider/pull/1014) fix: "failed to basic check approval msg"'s bug
* [#1016](https://github.com/zkMeLabs/mechain-storage-provider/pull/1016) fix:blocksyncer delete group bug
* [#996](https://github.com/zkMeLabs/mechain-storage-provider/pull/996) fix: db update bucket traffic by transaction
* [#955](https://github.com/zkMeLabs/mechain-storage-provider/pull/955) fix: fix src gvg is overwritten
* [#1018](https://github.com/zkMeLabs/mechain-storage-provider/pull/1018) fix: empty bucket when bucket migrate

TEST

* [#1001](https://github.com/zkMeLabs/mechain-storage-provider/pull/1001) test: sp db pkg adds unit test
* [#912](https://github.com/zkMeLabs/mechain-storage-provider/pull/912) ci: add coverage report for tests
* [#1007](https://github.com/zkMeLabs/mechain-storage-provider/pull/1007) test: sp db pkg adds unit test part II
* [#1009](https://github.com/zkMeLabs/mechain-storage-provider/pull/1009) test: core pkg generates mock files
* [#1019](https://github.com/zkMeLabs/mechain-storage-provider/pull/1019)test: gfspapp pkg adds UTs

## v0.2.4-alpha.1

FEATURES

* [#857](https://github.com/zkMeLabs/mechain-storage-provider/pull/857) feat: validate virtual group families' qualification
* [#985](https://github.com/zkMeLabs/mechain-storage-provider/pull/985) feat:time Ticker
* [#981](https://github.com/zkMeLabs/mechain-storage-provider/pull/981) feat: add tx confirm func and create virtual group retry
* [#968](https://github.com/zkMeLabs/mechain-storage-provider/pull/968) feat: bucket migrate check when load from db

REFACTOR

* [#983](https://github.com/zkMeLabs/mechain-storage-provider/pull/983) refine: refactor bucket migrate code
* [#953](https://github.com/zkMeLabs/mechain-storage-provider/pull/953) Refactor manager that dispatch task model
* [#976](https://github.com/zkMeLabs/mechain-storage-provider/pull/976) chore: refine migrate piece workflow
* [#960](https://github.com/zkMeLabs/mechain-storage-provider/pull/960) docs: polish sp docs to the lastest version

BUGFIX

* [#908](https://github.com/zkMeLabs/mechain-storage-provider/pull/908) fix: auth refactoring for security review
* [#987](https://github.com/zkMeLabs/mechain-storage-provider/pull/987) fix: ignore duplicate entry when create bucket traffic
* [#973](https://github.com/zkMeLabs/mechain-storage-provider/pull/973) fix: add missing path for pprof server
* [#966](https://github.com/zkMeLabs/mechain-storage-provider/pull/966) fix: refine migrate gvg task workflow
* [#965](https://github.com/zkMeLabs/mechain-storage-provider/pull/965) fix: repeated error xml msg
* [#963](https://github.com/zkMeLabs/mechain-storage-provider/pull/963) fix: only metadata and blocksyncer need to load bsdb
* [#955](https://github.com/zkMeLabs/mechain-storage-provider/pull/955) fix: check sp and bucket status when putting object
* [#935](https://github.com/zkMeLabs/mechain-storage-provider/pull/935) fix: empty bucket when bucket migrate
* [#927](https://github.com/zkMeLabs/mechain-storage-provider/pull/927) fix: filter complete migration buckets

TEST

* [#982](https://github.com/zkMeLabs/mechain-storage-provider/pull/982) test: piece store pkg storage dir adds unit test
* [#977](https://github.com/zkMeLabs/mechain-storage-provider/pull/977) test: piece store pkg adds unit test
* [#967](https://github.com/zkMeLabs/mechain-storage-provider/pull/967) test: package util adds unit test

## 0.2.3-alpha.11

FEATURES

* [#867](https://github.com/zkMeLabs/mechain-storage-provider/pull/867) feat: impl ListObjectsByGVGAndBucketForGC and object details
* [#888](https://github.com/zkMeLabs/mechain-storage-provider/pull/888) feat: metadata and block syncer monitor
* [#890](https://github.com/zkMeLabs/mechain-storage-provider/pull/890) feat: add generate gvg sp policy
* [#851](https://github.com/zkMeLabs/mechain-storage-provider/pull/851) feat: support query sp

REFACTOR

* [#860](https://github.com/zkMeLabs/mechain-storage-provider/pull/860) refactor: update quota consumption method

BUGFIX

* [#848](https://github.com/zkMeLabs/mechain-storage-provider/pull/848) fix: fix recover command
* [#876](https://github.com/zkMeLabs/mechain-storage-provider/pull/876) fix: GVGPickFilter CheckGVG func's paramter
* [#893](https://github.com/zkMeLabs/mechain-storage-provider/pull/893) fix:blocksyncer object map id
* [#868](https://github.com/zkMeLabs/mechain-storage-provider/pull/868) fix: fix aliyun credential expiration issue
* [#894](https://github.com/zkMeLabs/mechain-storage-provider/pull/894) feat: revert grpc keepalive params
* [#897](https://github.com/zkMeLabs/mechain-storage-provider/pull/897) fix: private universal endpoint special suffix handle
* [#898](https://github.com/zkMeLabs/mechain-storage-provider/pull/898) fix:blocksyncer monitor

## 0.2.3-alpha.7

FEATURES

* [#824](https://github.com/zkMeLabs/mechain-storage-provider/pull/824) feat: support sp exit and bucket migrate
* [#856](https://github.com/zkMeLabs/mechain-storage-provider/pull/856) feat: update local virtual group event
* [#853](https://github.com/zkMeLabs/mechain-storage-provider/pull/853) feat: update greenfield-go-sdk e2e version
* [#852](https://github.com/zkMeLabs/mechain-storage-provider/pull/852) ci: fix docker-ci.yml to push develop
* [#865](https://github.com/zkMeLabs/mechain-storage-provider/pull/865) feat: add bucket migrate & sp exit query cli

BUGFIX

* [#834](https://github.com/zkMeLabs/mechain-storage-provider/pull/834) fix: remove v2 Authorization
* [#832](https://github.com/zkMeLabs/mechain-storage-provider/pull/832) fix: add checking logic for sig length and public length
* [#839](https://github.com/zkMeLabs/mechain-storage-provider/pull/839) fix: blocksyncer panic
* [#847](https://github.com/zkMeLabs/mechain-storage-provider/pull/847) fix: block syncer copy object
* [#850](https://github.com/zkMeLabs/mechain-storage-provider/pull/850) fix: handle concurrent spdb table creation
* [#814](https://github.com/zkMeLabs/mechain-storage-provider/pull/814) fix: verify group permission
* [#858](https://github.com/zkMeLabs/mechain-storage-provider/pull/858) fix: resumable upload maxpayload size bugs
* [#863](https://github.com/zkMeLabs/mechain-storage-provider/pull/863) fix: optimize piece migration logic to avoid oom

## v0.2.3-alpha.2

FEATURES

* [#664](https://github.com/zkMeLabs/mechain-storage-provider/pull/664) feat: simulate discontinue transaction before broadcast
* [#643](https://github.com/zkMeLabs/mechain-storage-provider/pull/643) feat: customize http client using connection pool
* [#681](https://github.com/zkMeLabs/mechain-storage-provider/pull/681) feat: implement aliyun oss storage
* [#706](https://github.com/zkMeLabs/mechain-storage-provider/pull/706) feat: verify object permission by meta service
* [#699](https://github.com/zkMeLabs/mechain-storage-provider/pull/699) feat: SP database sharding
* [#795](https://github.com/zkMeLabs/mechain-storage-provider/pull/795) feat: basic workflow adaptation in sp exit

REFACTOR

* [#709](https://github.com/zkMeLabs/mechain-storage-provider/pull/709) refactor: manager dispatch task
* [#800](https://github.com/zkMeLabs/mechain-storage-provider/pull/800) refactor: async report task

BUGFIX

* [#672](https://github.com/zkMeLabs/mechain-storage-provider/pull/672) fix: fix data recovery
* [#690](https://github.com/zkMeLabs/mechain-storage-provider/pull/690) fix: re-enable the off chain auth api and add related ut
* [#810](https://github.com/zkMeLabs/mechain-storage-provider/pull/810) fix: fix aliyunfs by fetching credentials with AliCloud SDK
* [#808](https://github.com/zkMeLabs/mechain-storage-provider/pull/808) fix: fix authenticator
* [#817](https://github.com/zkMeLabs/mechain-storage-provider/pull/817) fix: resumable upload max payload size

## v0.2.3-alpha.1

FEATURES  

* [#638](https://github.com/zkMeLabs/mechain-storage-provider/pull/638) feat: support data recovery
* [#660](https://github.com/zkMeLabs/mechain-storage-provider/pull/660) feat: add download cache
* [#480](https://github.com/zkMeLabs/mechain-storage-provider/pull/480) feat: support resumable upload

REFACTOR

* [#649](https://github.com/zkMeLabs/mechain-storage-provider/pull/649) docs: sp docs add flowchart

BUGFIX

* [#648](https://github.com/zkMeLabs/mechain-storage-provider/pull/648) fix: request cannot be nil in latestBlockHeight

## v0.2.2-alpha.1

FEATURES

* [\#502](https://github.com/zkMeLabs/mechain-storage-provider/pull/502) feat: support b2 store
* [\#512](https://github.com/zkMeLabs/mechain-storage-provider/pull/512) feat: universal endpoint for private object
* [\#517](https://github.com/zkMeLabs/mechain-storage-provider/pull/517) feat:group add extra field
* [\#524](https://github.com/zkMeLabs/mechain-storage-provider/pull/524) feat: query storage params by timestamp
* [\#525](https://github.com/zkMeLabs/mechain-storage-provider/pull/525) feat: reject unseal object after upload or replicate fail
* [\#528](https://github.com/zkMeLabs/mechain-storage-provider/pull/528) feat: support loading tasks
* [\#530](https://github.com/zkMeLabs/mechain-storage-provider/pull/530) feat: add debug command
* [\#533](https://github.com/zkMeLabs/mechain-storage-provider/pull/533) feat: return repeated approval task
* [\#536](https://github.com/zkMeLabs/mechain-storage-provider/pull/536) feat:group add extra field
* [\#542](https://github.com/zkMeLabs/mechain-storage-provider/pull/542) feat: change get block height by ws protocol

REFACTOR

* [\#486](https://github.com/zkMeLabs/mechain-storage-provider/pull/486) refactor: off chain auth
* [\#493](https://github.com/zkMeLabs/mechain-storage-provider/pull/493) fix: refine gc object workflow
* [\#495](https://github.com/zkMeLabs/mechain-storage-provider/pull/495) perf: perf get object workflow
* [\#503](https://github.com/zkMeLabs/mechain-storage-provider/pull/503) fix: refine sp db update interface
* [\#515](https://github.com/zkMeLabs/mechain-storage-provider/pull/515) feat: refine get challenge info workflow
* [\#546](https://github.com/zkMeLabs/mechain-storage-provider/pull/546) docs: add sp infra deployment docs
* [\#557](https://github.com/zkMeLabs/mechain-storage-provider/pull/557) fix: refine error code in universal endpoint and auto-close the walle…

BUGFIX

* [\#487](https://github.com/zkMeLabs/mechain-storage-provider/pull/487) fix: init challenge task add storage params
* [\#499](https://github.com/zkMeLabs/mechain-storage-provider/pull/499) fix: permission api
* [\#509](https://github.com/zkMeLabs/mechain-storage-provider/pull/509) fix:blocksyncer oom

## v0.2.1-alpha.1

FEATURES

* [\#444](https://github.com/zkMeLabs/mechain-storage-provider/pull/444) feat: refactor v0.2.1 query cli
* [\#446](https://github.com/zkMeLabs/mechain-storage-provider/pull/446) feat: add p2p ant address config
* [\#449](https://github.com/zkMeLabs/mechain-storage-provider/pull/449) feat: metadata service and universal endpoint refactor v0.2.1
* [\#450](https://github.com/zkMeLabs/mechain-storage-provider/pull/450) refactor:blocksyncer
* [\#468](https://github.com/zkMeLabs/mechain-storage-provider/pull/468) feat: add error for cal nil model
* [\#471](https://github.com/zkMeLabs/mechain-storage-provider/pull/471) refactor: update listobjects & blocksyncer modules
* [\#473](https://github.com/zkMeLabs/mechain-storage-provider/pull/473) refactor: update stop serving module

BUGFIX

* [\#431](https://github.com/zkMeLabs/mechain-storage-provider/pull/431) fix: data query issues caused by character set replacement
* [\#439](https://github.com/zkMeLabs/mechain-storage-provider/pull/439) fix:blocksyncer oom
* [\#457](https://github.com/zkMeLabs/mechain-storage-provider/pull/457) fix: fix listobjects sql err
* [\#462](https://github.com/zkMeLabs/mechain-storage-provider/pull/462) fix: base app rcmgr span panic
* [\#464](https://github.com/zkMeLabs/mechain-storage-provider/pull/464) fix: task queue gc delay when call has method

## v0.2.0

FEATURES

* [\#358](https://github.com/zkMeLabs/mechain-storage-provider/pull/358) feat: sp services add pprof
* [\#379](https://github.com/zkMeLabs/mechain-storage-provider/pull/379) feat:block syncer add read concurrency support
* [\#383](https://github.com/zkMeLabs/mechain-storage-provider/pull/383) feat: add universal endpoint view option
* [\#389](https://github.com/zkMeLabs/mechain-storage-provider/pull/389) feat: signer async send sealObject tx
* [\#398](https://github.com/zkMeLabs/mechain-storage-provider/pull/398) feat: localup shell adds generate sp.info and db.info function
* [\#401](https://github.com/zkMeLabs/mechain-storage-provider/pull/401) feat: add dual db warm up support for blocksyncer
* [\#402](https://github.com/zkMeLabs/mechain-storage-provider/pull/402) feat: bsdb switch
* [\#404](https://github.com/zkMeLabs/mechain-storage-provider/pull/404) feat: list objects pagination & folder path
* [\#406](https://github.com/zkMeLabs/mechain-storage-provider/pull/406) feat: adapt greenfield v0.47
* [\#408](https://github.com/zkMeLabs/mechain-storage-provider/pull/408) feat: add gc worker
* [\#410](https://github.com/zkMeLabs/mechain-storage-provider/pull/410) feat: support full-memory replicate task
* [\#411](https://github.com/zkMeLabs/mechain-storage-provider/pull/411) feat:add upload download add bandwidth limit
* [\#412](https://github.com/zkMeLabs/mechain-storage-provider/pull/412) feat: add get object meta and get bucket meta apis

BUGFIX

* [\#355](https://github.com/zkMeLabs/mechain-storage-provider/pull/355) fix: universal endpoint spaces
* [\#360](https://github.com/zkMeLabs/mechain-storage-provider/pull/360) fix: sql parenthesis handling
* [\#378](https://github.com/zkMeLabs/mechain-storage-provider/pull/378) fix: support authv2 bucket-quota api
* [\#413](https://github.com/zkMeLabs/mechain-storage-provider/pull/413) fix: fix nil pointer and update db config

## v0.1.2

FEATURES

* [\#308](https://github.com/zkMeLabs/mechain-storage-provider/pull/308) feat: adds seal object metrics and refine some codes
* [\#313](https://github.com/zkMeLabs/mechain-storage-provider/pull/313) feat: verify permission api
* [\#314](https://github.com/zkMeLabs/mechain-storage-provider/pull/314) feat: support path-style api and add query upload progress api
* [\#318](https://github.com/zkMeLabs/mechain-storage-provider/pull/316) feat: update schema and order for list deleted objects
* [\#319](https://github.com/zkMeLabs/mechain-storage-provider/pull/319) feat: implement off-chain-auth solution
* [\#320](https://github.com/zkMeLabs/mechain-storage-provider/pull/320) chore: polish tests and docs
* [\#329](https://github.com/zkMeLabs/mechain-storage-provider/pull/329) feat: update greenfield to the latest version
* [\#338](https://github.com/zkMeLabs/mechain-storage-provider/pull/338) feat: block sycner add txhash when export events & juno version update
* [\#340](https://github.com/zkMeLabs/mechain-storage-provider/pull/340) feat: update metadata block syncer schema and add ListExpiredBucketsBySp
* [\#349](https://github.com/zkMeLabs/mechain-storage-provider/pull/349) fix: keep retrying when any blocksycner event handles failure

## v0.1.1

FEATURES

* [\#274](https://github.com/zkMeLabs/mechain-storage-provider/pull/274) feat: update stream record column names
* [\#275](https://github.com/zkMeLabs/mechain-storage-provider/pull/275) refactor: tasknode streaming process reduces memory usage
* [\#279](https://github.com/zkMeLabs/mechain-storage-provider/pull/283) feat: grpc client adds retry function
* [\#292](https://github.com/zkMeLabs/mechain-storage-provider/pull/292) feat: add table recreate func & block height metric for block sycner
* [\#295](https://github.com/zkMeLabs/mechain-storage-provider/pull/295) feat: support https protocol
* [\#296](https://github.com/zkMeLabs/mechain-storage-provider/pull/296) chore: change sqldb default config
* [\#299](https://github.com/zkMeLabs/mechain-storage-provider/pull/299) feat: add nat manager for p2p
* [\#304](https://github.com/zkMeLabs/mechain-storage-provider/pull/304) feat: support dns for p2p node
* [\#325](https://github.com/zkMeLabs/mechain-storage-provider/pull/325) feat: add universal endpoint
* [\#333](https://github.com/zkMeLabs/mechain-storage-provider/pull/333) fix: use EIP-4361 message template for off-chain-auth
* [\#339](https://github.com/zkMeLabs/mechain-storage-provider/pull/339) fix: permit anonymous users to access public object
* [\#347](https://github.com/zkMeLabs/mechain-storage-provider/pull/347) fix: add spdb and piece store metrics for downloader

BUGFIX

* [\#277](https://github.com/zkMeLabs/mechain-storage-provider/pull/277) fix: rcmgr leak for downloader service
* [\#278](https://github.com/zkMeLabs/mechain-storage-provider/pull/278) fix: uploader panic under db access error
* [\#279](https://github.com/zkMeLabs/mechain-storage-provider/pull/279) chore: change default rcmgr limit to no infinite
* [\#286](https://github.com/zkMeLabs/mechain-storage-provider/pull/286) fix: fix challenge memory is inaccurate
* [\#288](https://github.com/zkMeLabs/mechain-storage-provider/pull/288) fix: fix auth type v2 query object bug
* [\#306](https://github.com/zkMeLabs/mechain-storage-provider/pull/306) fix: fix multi update map bug and polish db error
* [\#337](https://github.com/zkMeLabs/mechain-storage-provider/pull/337) fix: permit anonymous users to access public objec

## v0.1.0

BUGFIX

* [\#258](https://github.com/zkMeLabs/mechain-storage-provider/pull/258) fix put object verify permission bug
* [\#264](https://github.com/zkMeLabs/mechain-storage-provider/pull/264) fix: fix payment apis nil pointer error
* [\#265](https://github.com/zkMeLabs/mechain-storage-provider/pull/265) fix: fix sa iam type to access s3
* [\#268](https://github.com/zkMeLabs/mechain-storage-provider/pull/268) feat: update buckets/objects order
* [\#270](https://github.com/zkMeLabs/mechain-storage-provider/pull/270) feat: update buckets/objects order
* [\#272](https://github.com/zkMeLabs/mechain-storage-provider/pull/272) fix: upgrade juno version for a property length fix

BUILD

* [\#259](https://github.com/zkMeLabs/mechain-storage-provider/pull/259) ci: fix release.yml uncorrect env var name
* [\#263](https://github.com/zkMeLabs/mechain-storage-provider/pull/263) feat: add e2e test to workflow

## v0.0.5

FEATURES

* [\#211](https://github.com/zkMeLabs/mechain-storage-provider/pull/211) feat: sp services add metrics
* [\#221](https://github.com/zkMeLabs/mechain-storage-provider/pull/221) feat: implement p2p protocol and rpc service
* [\#232](https://github.com/zkMeLabs/mechain-storage-provider/pull/232) chore: refine gRPC error code
* [\#235](https://github.com/zkMeLabs/mechain-storage-provider/pull/235) feat: implement metadata payment apis
* [\#244](https://github.com/zkMeLabs/mechain-storage-provider/pull/244) feat: update the juno version
* [\#246](https://github.com/zkMeLabs/mechain-storage-provider/pull/246) feat: resource manager

BUILD

* [\#231](https://github.com/zkMeLabs/mechain-storage-provider/pull/231) ci: add gosec checker

## v0.0.4

FEATURES

* [\#202](https://github.com/zkMeLabs/mechain-storage-provider/pull/202) feat: update get bucket apis
* [\#205](https://github.com/zkMeLabs/mechain-storage-provider/pull/205) fix: blocksyncer adapt event param to chain side and payment module added
* [\#206](https://github.com/zkMeLabs/mechain-storage-provider/pull/206) feat: support query quota and list read record
* [\#215](https://github.com/zkMeLabs/mechain-storage-provider/pull/215) fix: potential attack risks in on-chain storage module

IMPROVEMENT

* [\#188](https://github.com/zkMeLabs/mechain-storage-provider/pull/188) refactor: refactor metadata service
* [\#196](https://github.com/zkMeLabs/mechain-storage-provider/pull/196) docs: add sp docs
* [\#197](https://github.com/zkMeLabs/mechain-storage-provider/pull/197) refactor: rename stonenode, syncer to tasknode, recevier
* [\#200](https://github.com/zkMeLabs/mechain-storage-provider/pull/200) docs: refining readme
* [\#208](https://github.com/zkMeLabs/mechain-storage-provider/pull/208) docs: add block syncer config
* [\#209](https://github.com/zkMeLabs/mechain-storage-provider/pull/209) fix: block syncer db response style

BUGFIX

* [\#189](https://github.com/zkMeLabs/mechain-storage-provider/pull/189) fix: fix approval expired height bug
* [\#212](https://github.com/zkMeLabs/mechain-storage-provider/pull/212) fix: authv2 workflow
* [\#216](https://github.com/zkMeLabs/mechain-storage-provider/pull/216) fix: metadata buckets api

BUILD

* [\#179](https://github.com/zkMeLabs/mechain-storage-provider/pull/179) ci: add branch naming rules
* [\#198](https://github.com/zkMeLabs/mechain-storage-provider/pull/198) build: replace go1.19 with go1.18

## v0.0.3

FEATURES

* [\#169](https://github.com/zkMeLabs/mechain-storage-provider/pull/169) feat: piece store adds minio storage type
* [\#172](https://github.com/zkMeLabs/mechain-storage-provider/pull/172) feat: implement manager module
* [\#173](https://github.com/zkMeLabs/mechain-storage-provider/pull/173) feat: add check billing

IMPROVEMENT

* [\#154](https://github.com/zkMeLabs/mechain-storage-provider/pull/154) feat: syncer opt with chain data struct
* [\#156](https://github.com/zkMeLabs/mechain-storage-provider/pull/156) refactor: implement sp db, remove meta db and job db
* [\#157](https://github.com/zkMeLabs/mechain-storage-provider/pull/157) refactor: polish gateway module
* [\#162](https://github.com/zkMeLabs/mechain-storage-provider/pull/162) feat: add command for devops and config log
* [\#165](https://github.com/zkMeLabs/mechain-storage-provider/pull/165) feat: improve sync piece efficiency
* [\#171](https://github.com/zkMeLabs/mechain-storage-provider/pull/171) feat: add localup script

## v0.0.2

This release includes following features:

1. Implement the connection with the greenfield chain, and the upload and download of payload, including basic permission verification.
2. Implement the signer service for storage providers to sign the on-chain transactions.
3. Implement the communication of HTTP between SPs instead of gRPC.

* [\#131](https://github.com/zkMeLabs/mechain-storage-provider/pull/131) feat: add chain client to sp
* [\#119](https://github.com/zkMeLabs/mechain-storage-provider/pull/119) feat: implement signer service
* [\#128](https://github.com/zkMeLabs/mechain-storage-provider/pull/128) feat: stone node sends piece data to gateway
* [\#127](https://github.com/zkMeLabs/mechain-storage-provider/pull/127) feat: implement gateway challenge workflow
* [\#133](https://github.com/zkMeLabs/mechain-storage-provider/pull/133) fix: upgrade greenfield version to fix the signing bug
* [\#130](https://github.com/zkMeLabs/mechain-storage-provider/pull/130) fix: use env var to get bucket url

## v0.0.1

IMPROVEMENT

* [\#65](https://github.com/zkMeLabs/mechain-storage-provider/pull/65) feat: gateway add verify signature
* [\#43](https://github.com/zkMeLabs/mechain-storage-provider/pull/43) feat(uploader): add getAuth interface
* [\#68](https://github.com/zkMeLabs/mechain-storage-provider/pull/68) refactor: add jobdb v2 interface, objectID as primary key
* [\#70](https://github.com/zkMeLabs/mechain-storage-provider/pull/70) feat: change index from create object hash to object id
* [\#73](https://github.com/zkMeLabs/mechain-storage-provider/pull/73) feat(metadb): add sql metadb
* [\#82](https://github.com/zkMeLabs/mechain-storage-provider/pull/82) feat(stone_node): supports sending data to different storage provider
* [\#66](https://github.com/zkMeLabs/mechain-storage-provider/pull/66) fix: adjust the dispatching strategy of replica and inline data into storage provider
* [\#69](https://github.com/zkMeLabs/mechain-storage-provider/pull/69) fix: use multi-dimensional array to send piece data and piece hash
* [\#101](https://github.com/zkMeLabs/mechain-storage-provider/pull/101) fix: remove tokens from config and use env vars to load tokens
* [\#83](https://github.com/zkMeLabs/mechain-storage-provider/pull/83) chore(sql): polish sql workflow
* [\#87](https://github.com/zkMeLabs/mechain-storage-provider/pull/87) chore: add setup-test-env tool

BUILD

* [\#74](https://github.com/zkMeLabs/mechain-storage-provider/pull/74) ci: add docker release pipe
* [\#67](https://github.com/zkMeLabs/mechain-storage-provider/pull/67) ci: add commit lint, code lint and unit test ci files
* [\#85](https://github.com/zkMeLabs/mechain-storage-provider/pull/85) chore: add pull request template
* [\#105](https://github.com/zkMeLabs/mechain-storage-provider/pull/105) fix: add release action

## v0.0.1-alpha

This release includes features, mainly:

1. Implement the upload and download of payload data and the challenge handler api of piece data;
2. Implement the main architecture of greenfield storage provider:  
   2.1 gateway: the entry point of each sp, parses requests from the client and dispatches them to special service;  
   2.2 uploader: receives the object's payload data, splits it into segments, and stores them in piece store;
   2.3 downloader: handles the user's downloading request and gets object data from the piece store;
   2.4 stonehub: works as state machine to handle all background jobs, each job includes several tasks;
   2.5 stonenode: works as the execute unit, it watches the stonehub tasks(the smallest unit of a job) and executes them;
   2.6 syncer: receives data pieces from primary sp and stores them in the piece store when sp works as a secondary sp;
3. Implement one-click deployment and one-click running test, which is convenient for developers and testers to experience the mechain-sp.

* [\#7](https://github.com/zkMeLabs/mechain-storage-provider/pull/7) feat(gateway/uploader): add gateway and uploader skeleton
* [\#16](https://github.com/zkMeLabs/mechain-storage-provider/pull/16) Add secondary syncer service
* [\#17](https://github.com/zkMeLabs/mechain-storage-provider/pull/17) feat: implement of upload payload in stone hub side
* [\#29](https://github.com/zkMeLabs/mechain-storage-provider/pull/28) fix: ston node goroutine model
* [\#38](https://github.com/zkMeLabs/mechain-storage-provider/pull/38) feat: implement the challenge service
* [\#9](https://github.com/zkMeLabs/mechain-storage-provider/pull/9) add service lifecycle module
* [\#2](https://github.com/zkMeLabs/mechain-storage-provider/pull/2) add piecestore module
* [\#18](https://github.com/zkMeLabs/mechain-storage-provider/pull/18) feat: add job meta orm
* [\#60](https://github.com/zkMeLabs/mechain-storage-provider/pull/60) test: add run cases
