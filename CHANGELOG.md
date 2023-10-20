# Changelog

## v1.0.1

FEATURES
* [#1181](https://github.com/bnb-chain/greenfield-storage-provider/pull/1181) feat: api rate limiter refactor
* [#1188](https://github.com/bnb-chain/greenfield-storage-provider/pull/1188) feat:blocksyncer fix data command

BIGFIX
* [#1186](https://github.com/bnb-chain/greenfield-storage-provider/pull/1186) fix:batch delete group member
* [#1187](https://github.com/bnb-chain/greenfield-storage-provider/pull/1187) fix:payment refundable update

## v1.0.0
This is the first official version for the main-net deployment.

## v0.2.6-hf.2
BUGFIX
* [#1171](https://github.com/bnb-chain/greenfield-storage-provider/pull/1171) fix: fix init the NotAvailableSpIdx

## v0.2.6-hf.1

BUGFIX
* [#1166](https://github.com/bnb-chain/greenfield-storage-provider/pull/1166) fix: delete group fix job
* [#1160](https://github.com/bnb-chain/greenfield-storage-provider/pull/1160) fix: fix replication failed SP err recording issue
* [#1159](https://github.com/bnb-chain/greenfield-storage-provider/pull/1159) fix: list objects by gvg sql bug
* [#1165](https://github.com/bnb-chain/greenfield-storage-provider/pull/1165) fix: verify permission 500 code and update UT
* [#1155](https://github.com/bnb-chain/greenfield-storage-provider/pull/1155) fix: fix manager duplicate entry when inserting upload progress
* [#1158](https://github.com/bnb-chain/greenfield-storage-provider/pull/1158) fix: add xml response for rate limit error
* [#1156](https://github.com/bnb-chain/greenfield-storage-provider/pull/1156) fix: recovery object status should be sealed
* [#1154](https://github.com/bnb-chain/greenfield-storage-provider/pull/1154) fix: add ut for universal endpoint


## v0.2.6

TEST
* [#1157](https://github.com/bnb-chain/greenfield-storage-provider/pull/1157) test: add metadata api UT

## v0.2.6-alpha.2

BUGFIX
* [#1151](https://github.com/bnb-chain/greenfield-storage-provider/pull/1151) fix: update new logic of metadata apis

## v0.2.6-alpha.1

FEATURES
* [#1109](https://github.com/bnb-chain/greenfield-storage-provider/pull/1109) feat: add migrate piece auth check
* [#1131](https://github.com/bnb-chain/greenfield-storage-provider/pull/1131) feat: limit creating object and migrating bucket approval

BUGFIX
* [#1130](https://github.com/bnb-chain/greenfield-storage-provider/pull/1130) fix: defer returning err until all goroutines are done
* [#1134](https://github.com/bnb-chain/greenfield-storage-provider/pull/1134) fix: readme
* [#1139](https://github.com/bnb-chain/greenfield-storage-provider/pull/1139) fix: refactoring code for getObject and Universal apis
* [#1140](https://github.com/bnb-chain/greenfield-storage-provider/pull/1140) fix: add replicate check permission
* [#1142](https://github.com/bnb-chain/greenfield-storage-provider/pull/1142) fix: resumable queue bug
* [#1145](https://github.com/bnb-chain/greenfield-storage-provider/pull/1145) fix: policy statement is empty
* [#1146](https://github.com/bnb-chain/greenfield-storage-provider/pull/1146) fix: fix integrity hash check
* [#1147](https://github.com/bnb-chain/greenfield-storage-provider/pull/1147) fix: fix uploader server read closed channel

TEST
* [#1143](https://github.com/bnb-chain/greenfield-storage-provider/pull/1143) test: add blocksyncer ut case

## v0.2.5

BUGFIX
* [#1135](https://github.com/bnb-chain/greenfield-storage-provider/pull/1135) fix: effect allow issue

## v0.2.5-alpha.3

FEATURES
* [#1129](https://github.com/bnb-chain/greenfield-storage-provider/pull/1129) feat: add back metadata go routine listener
* [#1128](https://github.com/bnb-chain/greenfield-storage-provider/pull/1128) feat: support expire time check of SP task info

BUGFIX
* [#1092](https://github.com/bnb-chain/greenfield-storage-provider/pull/1092) fix: support reimburse quota if download consumed extra quota
* [#1127](https://github.com/bnb-chain/greenfield-storage-provider/pull/1127) fix:bs groups unique index
* [#1125](https://github.com/bnb-chain/greenfield-storage-provider/pull/1125) fix: expiration time for statement

TEST
* [#1126](https://github.com/bnb-chain/greenfield-storage-provider/pull/1126) test: modular/executor pkg adds UTs

## v0.2.5-alpha.2

FEATURES
* [#1112](https://github.com/bnb-chain/greenfield-storage-provider/pull/1112) feat: support GfSpGetSPMigrateBucketNumber

BUGFIX
* [#1121](https://github.com/bnb-chain/greenfield-storage-provider/pull/1121) fix: fix tls MinVersion to 1.2
* [#1119](https://github.com/bnb-chain/greenfield-storage-provider/pull/1119) fix: stop all replicate jobs and done replicate job by context
* [#1117](https://github.com/bnb-chain/greenfield-storage-provider/pull/1117) fix: include private cause bucket not found

TEST
* [#1115](https://github.com/bnb-chain/greenfield-storage-provider/pull/1115) test: modular/gater pkg adds UTs part II

## v0.2.5-alpha.1

FEATURES
* [#1029](https://github.com/bnb-chain/greenfield-storage-provider/pull/1029) feat: use aliyun oss sdk to visit oss
* [#1042](https://github.com/bnb-chain/greenfield-storage-provider/pull/1042) feat: add go routine metrics
* [#1057](https://github.com/bnb-chain/greenfield-storage-provider/pull/1057) feat: list object policies and add number of group members
* [#1067](https://github.com/bnb-chain/greenfield-storage-provider/pull/1067) feat: Docker image distroless update
* [#1111](https://github.com/bnb-chain/greenfield-storage-provider/pull/1111) feat: only persist the init off-chain-auth record when first updating
* [#1103](https://github.com/bnb-chain/greenfield-storage-provider/pull/1103) feat: list groups by ids
* [#1088](https://github.com/bnb-chain/greenfield-storage-provider/pull/1088) feat: recover object list
* [#1077](https://github.com/bnb-chain/greenfield-storage-provider/pull/1077) feat: support ListUserPaymentAccounts & ListPaymentAccountStreams

BUGFIX
* [#1039](https://github.com/bnb-chain/greenfield-storage-provider/pull/1039) fix: remove funding key in singer module
* [#1053](https://github.com/bnb-chain/greenfield-storage-provider/pull/1053) fix: add more metrics log for PickUpTask
* [#1054](https://github.com/bnb-chain/greenfield-storage-provider/pull/1054) fix: fix quota db to support get quota by month
* [#1062](https://github.com/bnb-chain/greenfield-storage-provider/pull/1062) fix: add fingerprint to approval key
* [#1071](https://github.com/bnb-chain/greenfield-storage-provider/pull/1071) fix: cancel migrate bucket bug
* [#1074](https://github.com/bnb-chain/greenfield-storage-provider/pull/1074) fix: set correct Content-Disposition when downloading an object
* [#1104](https://github.com/bnb-chain/greenfield-storage-provider/pull/1104) fix: refactor quota  table
* [#1107](https://github.com/bnb-chain/greenfield-storage-provider/pull/1107) fix: api rate limiter path sequence
* [#1106](https://github.com/bnb-chain/greenfield-storage-provider/pull/1106) fix: sec issue about conversion alerts
* [#1089](https://github.com/bnb-chain/greenfield-storage-provider/pull/1089) fix: sec issue about conversion alerts
* [#1098](https://github.com/bnb-chain/greenfield-storage-provider/pull/1098) fix: self sp id retrieval
* [#1097](https://github.com/bnb-chain/greenfield-storage-provider/pull/1097) fix: fix group ExpirationTime bug
* [#1093](https://github.com/bnb-chain/greenfield-storage-provider/pull/1093) fix: verify permission if expiration time = 0 bug
* [#1087](https://github.com/bnb-chain/greenfield-storage-provider/pull/1087) fix: upgrade libp2p and cosmos-sdk version to solve security issues
* [#1084](https://github.com/bnb-chain/greenfield-storage-provider/pull/1084) fix: resumable upload support 64g file
* [#776](https://github.com/bnb-chain/greenfield-storage-provider/pull/776) fix: code security check
* [#1078](https://github.com/bnb-chain/greenfield-storage-provider/pull/1078) fix: add more resource manager log

TEST
* [#1032](https://github.com/bnb-chain/greenfield-storage-provider/pull/1032) test: modular/approver pkg adds UTs
* [#1035](https://github.com/bnb-chain/greenfield-storage-provider/pull/1035) test: metadata bsdb UT
* [#1040](https://github.com/bnb-chain/greenfield-storage-provider/pull/1040) test: add cmd ut
* [#1046](https://github.com/bnb-chain/greenfield-storage-provider/pull/1046) test: add downloader ut
* [#1068](https://github.com/bnb-chain/greenfield-storage-provider/pull/1068) test: gfspapp pkg adds UTs part II
* [#1073](https://github.com/bnb-chain/greenfield-storage-provider/pull/1073) test: gfspconfig and gfsppieceop pkg add UTs
* [#1108](https://github.com/bnb-chain/greenfield-storage-provider/pull/1108) test: modular/gater pkg adds UTs
* [#1095](https://github.com/bnb-chain/greenfield-storage-provider/pull/1095) test: modular/uploader pkg adds UTs
* [#1096](https://github.com/bnb-chain/greenfield-storage-provider/pull/1096) test: base/gfsprcmgr pkg adds UTs
* [#1090](https://github.com/bnb-chain/greenfield-storage-provider/pull/1090) test: base/gfspclient pkg adds UTs
* [#1082](https://github.com/bnb-chain/greenfield-storage-provider/pull/1082) test: base/types directory adds UTs
* [#1076](https://github.com/bnb-chain/greenfield-storage-provider/pull/1076) test: gfsptqueue and gfspvgmgr pkg add UTs
* [#1060](https://github.com/bnb-chain/greenfield-storage-provider/pull/1060) test: add receiver ut
* [#1086](https://github.com/bnb-chain/greenfield-storage-provider/pull/1086) test: add blocksyncer e2e test
 
## v0.2.4-alpha.9

FEATURES
* [#989](https://github.com/bnb-chain/greenfield-storage-provider/pull/989) feat: impl group apis and fix verify permission bug
* [#1008](https://github.com/bnb-chain/greenfield-storage-provider/pull/1008) feat: change auth api response from json to xml
* [#1010](https://github.com/bnb-chain/greenfield-storage-provider/pull/1010) feat:blocksyncer add realtime mode
* [#1015](https://github.com/bnb-chain/greenfield-storage-provider/pull/1015) feat: retrieve groups where the user is the owner
* [#1012](https://github.com/bnb-chain/greenfield-storage-provider/pull/1012) feat: error handle updates to provide useful messages
* [#1025](https://github.com/bnb-chain/greenfield-storage-provider/pull/1025) feat: change json response body to xml

BUGFIX
* [#999](https://github.com/bnb-chain/greenfield-storage-provider/pull/999) fix: fix bug for metadata crash
* [#1000](https://github.com/bnb-chain/greenfield-storage-provider/pull/1000) fix: fix occasional compile error
* [#1006](https://github.com/bnb-chain/greenfield-storage-provider/pull/1006) fix: rename api name and replace post to get
* [#1014](https://github.com/bnb-chain/greenfield-storage-provider/pull/1014) fix: "failed to basic check approval msg"'s bug
* [#1016](https://github.com/bnb-chain/greenfield-storage-provider/pull/1016) fix:blocksyncer delete group bug
* [#996](https://github.com/bnb-chain/greenfield-storage-provider/pull/996) fix: db update bucket traffic by transaction
* [#955](https://github.com/bnb-chain/greenfield-storage-provider/pull/955) fix: fix src gvg is overwritten
* [#1018](https://github.com/bnb-chain/greenfield-storage-provider/pull/1018) fix: empty bucket when bucket migrate

TEST
* [#1001](https://github.com/bnb-chain/greenfield-storage-provider/pull/1001) test: sp db pkg adds unit test
* [#912](https://github.com/bnb-chain/greenfield-storage-provider/pull/912) ci: add coverage report for tests
* [#1007](https://github.com/bnb-chain/greenfield-storage-provider/pull/1007) test: sp db pkg adds unit test part II
* [#1009](https://github.com/bnb-chain/greenfield-storage-provider/pull/1009) test: core pkg generates mock files
* [#1019](https://github.com/bnb-chain/greenfield-storage-provider/pull/1019)test: gfspapp pkg adds UTs

## v0.2.4-alpha.1

FEATURES
* [#857](https://github.com/bnb-chain/greenfield-storage-provider/pull/857) feat: validate virtual group families' qualification
* [#985](https://github.com/bnb-chain/greenfield-storage-provider/pull/985) feat:time Ticker
* [#981](https://github.com/bnb-chain/greenfield-storage-provider/pull/981) feat: add tx confirm func and create virtual group retry
* [#968](https://github.com/bnb-chain/greenfield-storage-provider/pull/968) feat: bucket migrate check when load from db


REFACTOR
* [#983](https://github.com/bnb-chain/greenfield-storage-provider/pull/983) refine: refactor bucket migrate code
* [#953](https://github.com/bnb-chain/greenfield-storage-provider/pull/953) Refactor manager that dispatch task model
* [#976](https://github.com/bnb-chain/greenfield-storage-provider/pull/976) chore: refine migrate piece workflow
* [#960](https://github.com/bnb-chain/greenfield-storage-provider/pull/960) docs: polish sp docs to the lastest version

BUGFIX
* [#908](https://github.com/bnb-chain/greenfield-storage-provider/pull/908) fix: auth refactoring for security review
* [#987](https://github.com/bnb-chain/greenfield-storage-provider/pull/987) fix: ignore duplicate entry when create bucket traffic
* [#973](https://github.com/bnb-chain/greenfield-storage-provider/pull/973) fix: add missing path for pprof server
* [#966](https://github.com/bnb-chain/greenfield-storage-provider/pull/966) fix: refine migrate gvg task workflow
* [#965](https://github.com/bnb-chain/greenfield-storage-provider/pull/965) fix: repeated error xml msg
* [#963](https://github.com/bnb-chain/greenfield-storage-provider/pull/963) fix: only metadata and blocksyncer need to load bsdb
* [#955](https://github.com/bnb-chain/greenfield-storage-provider/pull/955) fix: check sp and bucket status when putting object
* [#935](https://github.com/bnb-chain/greenfield-storage-provider/pull/935) fix: empty bucket when bucket migrate
* [#927](https://github.com/bnb-chain/greenfield-storage-provider/pull/927) fix: filter complete migration buckets

TEST
* [#982](https://github.com/bnb-chain/greenfield-storage-provider/pull/982) test: piece store pkg storage dir adds unit test
* [#977](https://github.com/bnb-chain/greenfield-storage-provider/pull/977) test: piece store pkg adds unit test
* [#967](https://github.com/bnb-chain/greenfield-storage-provider/pull/967) test: package util adds unit test

## 0.2.3-alpha.11

FEATURES
* [#867](https://github.com/bnb-chain/greenfield-storage-provider/pull/867) feat: impl ListObjectsByGVGAndBucketForGC and object details
* [#888](https://github.com/bnb-chain/greenfield-storage-provider/pull/888) feat: metadata and block syncer monitor
* [#890](https://github.com/bnb-chain/greenfield-storage-provider/pull/890) feat: add generate gvg sp policy
* [#851](https://github.com/bnb-chain/greenfield-storage-provider/pull/851) feat: support query sp

REFACTOR
* [#860](https://github.com/bnb-chain/greenfield-storage-provider/pull/860) refactor: update quota consumption method

BUGFIX
* [#848](https://github.com/bnb-chain/greenfield-storage-provider/pull/848) fix: fix recover command
* [#876](https://github.com/bnb-chain/greenfield-storage-provider/pull/876) fix: GVGPickFilter CheckGVG func's paramter
* [#893](https://github.com/bnb-chain/greenfield-storage-provider/pull/893) fix:blocksyncer object map id
* [#868](https://github.com/bnb-chain/greenfield-storage-provider/pull/868) fix: fix aliyun credential expiration issue
* [#894](https://github.com/bnb-chain/greenfield-storage-provider/pull/894) feat: revert grpc keepalive params
* [#897](https://github.com/bnb-chain/greenfield-storage-provider/pull/897) fix: private universal endpoint special suffix handle
* [#898](https://github.com/bnb-chain/greenfield-storage-provider/pull/898) fix:blocksyncer monitor


## 0.2.3-alpha.7

FEATURES
* [#824](https://github.com/bnb-chain/greenfield-storage-provider/pull/824) feat: support sp exit and bucket migrate
* [#856](https://github.com/bnb-chain/greenfield-storage-provider/pull/856) feat: update local virtual group event
* [#853](https://github.com/bnb-chain/greenfield-storage-provider/pull/853) feat: update greenfield-go-sdk e2e version
* [#852](https://github.com/bnb-chain/greenfield-storage-provider/pull/852) ci: fix docker-ci.yml to push develop
* [#865](https://github.com/bnb-chain/greenfield-storage-provider/pull/865) feat: add bucket migrate & sp exit query cli

BUGFIX
* [#834](https://github.com/bnb-chain/greenfield-storage-provider/pull/834) fix: remove v2 Authorization
* [#832](https://github.com/bnb-chain/greenfield-storage-provider/pull/832) fix: add checking logic for sig length and public length
* [#839](https://github.com/bnb-chain/greenfield-storage-provider/pull/839) fix: blocksyncer panic
* [#847](https://github.com/bnb-chain/greenfield-storage-provider/pull/847) fix: block syncer copy object
* [#850](https://github.com/bnb-chain/greenfield-storage-provider/pull/850) fix: handle concurrent spdb table creation
* [#814](https://github.com/bnb-chain/greenfield-storage-provider/pull/814) fix: verify group permission
* [#858](https://github.com/bnb-chain/greenfield-storage-provider/pull/858) fix: resumable upload maxpayload size bugs
* [#863](https://github.com/bnb-chain/greenfield-storage-provider/pull/863) fix: optimize piece migration logic to avoid oom
 

## v0.2.3-alpha.2
FEATURES
* [#664](https://github.com/bnb-chain/greenfield-storage-provider/pull/664) feat: simulate discontinue transaction before broadcast
* [#643](https://github.com/bnb-chain/greenfield-storage-provider/pull/643) feat: customize http client using connection pool
* [#681](https://github.com/bnb-chain/greenfield-storage-provider/pull/681) feat: implement aliyun oss storage
* [#706](https://github.com/bnb-chain/greenfield-storage-provider/pull/706) feat: verify object permission by meta service
* [#699](https://github.com/bnb-chain/greenfield-storage-provider/pull/699) feat: SP database sharding
* [#795](https://github.com/bnb-chain/greenfield-storage-provider/pull/795) feat: basic workflow adaptation in sp exit

REFACTOR
* [#709](https://github.com/bnb-chain/greenfield-storage-provider/pull/709) refactor: manager dispatch task
* [#800](https://github.com/bnb-chain/greenfield-storage-provider/pull/800) refactor: async report task

BUGFIX
* [#672](https://github.com/bnb-chain/greenfield-storage-provider/pull/672) fix: fix data recovery
* [#690](https://github.com/bnb-chain/greenfield-storage-provider/pull/690) fix: re-enable the off chain auth api and add related ut
* [#810](https://github.com/bnb-chain/greenfield-storage-provider/pull/810) fix: fix aliyunfs by fetching credentials with AliCloud SDK
* [#808](https://github.com/bnb-chain/greenfield-storage-provider/pull/808) fix: fix authenticator
* [#817](https://github.com/bnb-chain/greenfield-storage-provider/pull/817) fix: resumable upload max payload size

## v0.2.3-alpha.1

FEATURES  
* [#638](https://github.com/bnb-chain/greenfield-storage-provider/pull/638) feat: support data recovery
* [#660](https://github.com/bnb-chain/greenfield-storage-provider/pull/660) feat: add download cache 
* [#480](https://github.com/bnb-chain/greenfield-storage-provider/pull/480) feat: support resumable upload

REFACTOR
* [#649](https://github.com/bnb-chain/greenfield-storage-provider/pull/649) docs: sp docs add flowchart

BUGFIX
* [#648](https://github.com/bnb-chain/greenfield-storage-provider/pull/648) fix: request cannot be nil in latestBlockHeight


## v0.2.2-alpha.1

FEATURES
* [\#502](https://github.com/bnb-chain/greenfield-storage-provider/pull/502) feat: support b2 store
* [\#512](https://github.com/bnb-chain/greenfield-storage-provider/pull/512) feat: universal endpoint for private object
* [\#517](https://github.com/bnb-chain/greenfield-storage-provider/pull/517) feat:group add extra field
* [\#524](https://github.com/bnb-chain/greenfield-storage-provider/pull/524) feat: query storage params by timestamp
* [\#525](https://github.com/bnb-chain/greenfield-storage-provider/pull/525) feat: reject unseal object after upload or replicate fail
* [\#528](https://github.com/bnb-chain/greenfield-storage-provider/pull/528) feat: support loading tasks
* [\#530](https://github.com/bnb-chain/greenfield-storage-provider/pull/530) feat: add debug command
* [\#533](https://github.com/bnb-chain/greenfield-storage-provider/pull/533) feat: return repeated approval task
* [\#536](https://github.com/bnb-chain/greenfield-storage-provider/pull/536) feat:group add extra field
* [\#542](https://github.com/bnb-chain/greenfield-storage-provider/pull/542) feat: change get block height by ws protocol

REFACTOR
* [\#486](https://github.com/bnb-chain/greenfield-storage-provider/pull/486) refactor: off chain auth
* [\#493](https://github.com/bnb-chain/greenfield-storage-provider/pull/493) fix: refine gc object workflow
* [\#495](https://github.com/bnb-chain/greenfield-storage-provider/pull/495) perf: perf get object workflow
* [\#503](https://github.com/bnb-chain/greenfield-storage-provider/pull/503) fix: refine sp db update interface
* [\#515](https://github.com/bnb-chain/greenfield-storage-provider/pull/515) feat: refine get challenge info workflow
* [\#546](https://github.com/bnb-chain/greenfield-storage-provider/pull/546) docs: add sp infra deployment docs
* [\#557](https://github.com/bnb-chain/greenfield-storage-provider/pull/557) fix: refine error code in universal endpoint and auto-close the walle…

BUGFIX
* [\#487](https://github.com/bnb-chain/greenfield-storage-provider/pull/487) fix: init challenge task add storage params
* [\#499](https://github.com/bnb-chain/greenfield-storage-provider/pull/499) fix: permission api
* [\#509](https://github.com/bnb-chain/greenfield-storage-provider/pull/509) fix:blocksyncer oom

## v0.2.1-alpha.1

FEATURES
* [\#444](https://github.com/bnb-chain/greenfield-storage-provider/pull/444) feat: refactor v0.2.1 query cli
* [\#446](https://github.com/bnb-chain/greenfield-storage-provider/pull/446) feat: add p2p ant address config
* [\#449](https://github.com/bnb-chain/greenfield-storage-provider/pull/449) feat: metadata service and universal endpoint refactor v0.2.1
* [\#450](https://github.com/bnb-chain/greenfield-storage-provider/pull/450) refactor:blocksyncer
* [\#468](https://github.com/bnb-chain/greenfield-storage-provider/pull/468) feat: add error for cal nil model
* [\#471](https://github.com/bnb-chain/greenfield-storage-provider/pull/471) refactor: update listobjects & blocksyncer modules
* [\#473](https://github.com/bnb-chain/greenfield-storage-provider/pull/473) refactor: update stop serving module

BUGFIX
* [\#431](https://github.com/bnb-chain/greenfield-storage-provider/pull/431) fix: data query issues caused by character set replacement
* [\#439](https://github.com/bnb-chain/greenfield-storage-provider/pull/439) fix:blocksyncer oom
* [\#457](https://github.com/bnb-chain/greenfield-storage-provider/pull/457) fix: fix listobjects sql err
* [\#462](https://github.com/bnb-chain/greenfield-storage-provider/pull/462) fix: base app rcmgr span panic
* [\#464](https://github.com/bnb-chain/greenfield-storage-provider/pull/464) fix: task queue gc delay when call has method

## v0.2.0

FEATURES
* [\#358](https://github.com/bnb-chain/greenfield-storage-provider/pull/358) feat: sp services add pprof
* [\#379](https://github.com/bnb-chain/greenfield-storage-provider/pull/379) feat:block syncer add read concurrency support
* [\#383](https://github.com/bnb-chain/greenfield-storage-provider/pull/383) feat: add universal endpoint view option
* [\#389](https://github.com/bnb-chain/greenfield-storage-provider/pull/389) feat: signer async send sealObject tx
* [\#398](https://github.com/bnb-chain/greenfield-storage-provider/pull/398) feat: localup shell adds generate sp.info and db.info function
* [\#401](https://github.com/bnb-chain/greenfield-storage-provider/pull/401) feat: add dual db warm up support for blocksyncer
* [\#402](https://github.com/bnb-chain/greenfield-storage-provider/pull/402) feat: bsdb switch
* [\#404](https://github.com/bnb-chain/greenfield-storage-provider/pull/404) feat: list objects pagination & folder path
* [\#406](https://github.com/bnb-chain/greenfield-storage-provider/pull/406) feat: adapt greenfield v0.47
* [\#408](https://github.com/bnb-chain/greenfield-storage-provider/pull/408) feat: add gc worker
* [\#410](https://github.com/bnb-chain/greenfield-storage-provider/pull/410) feat: support full-memory replicate task
* [\#411](https://github.com/bnb-chain/greenfield-storage-provider/pull/411) feat:add upload download add bandwidth limit
* [\#412](https://github.com/bnb-chain/greenfield-storage-provider/pull/412) feat: add get object meta and get bucket meta apis

BUGFIX
* [\#355](https://github.com/bnb-chain/greenfield-storage-provider/pull/355) fix: universal endpoint spaces
* [\#360](https://github.com/bnb-chain/greenfield-storage-provider/pull/360) fix: sql parenthesis handling
* [\#378](https://github.com/bnb-chain/greenfield-storage-provider/pull/378) fix: support authv2 bucket-quota api
* [\#413](https://github.com/bnb-chain/greenfield-storage-provider/pull/413) fix: fix nil pointer and update db config


## v0.1.2

FEATURES
* [\#308](https://github.com/bnb-chain/greenfield-storage-provider/pull/308) feat: adds seal object metrics and refine some codes
* [\#313](https://github.com/bnb-chain/greenfield-storage-provider/pull/313) feat: verify permission api
* [\#314](https://github.com/bnb-chain/greenfield-storage-provider/pull/314) feat: support path-style api and add query upload progress api
* [\#318](https://github.com/bnb-chain/greenfield-storage-provider/pull/316) feat: update schema and order for list deleted objects
* [\#319](https://github.com/bnb-chain/greenfield-storage-provider/pull/319) feat: implement off-chain-auth solution
* [\#320](https://github.com/bnb-chain/greenfield-storage-provider/pull/320) chore: polish tests and docs
* [\#329](https://github.com/bnb-chain/greenfield-storage-provider/pull/329) feat: update greenfield to the latest version
* [\#338](https://github.com/bnb-chain/greenfield-storage-provider/pull/338) feat: block sycner add txhash when export events & juno version update
* [\#340](https://github.com/bnb-chain/greenfield-storage-provider/pull/340) feat: update metadata block syncer schema and add ListExpiredBucketsBySp
* [\#349](https://github.com/bnb-chain/greenfield-storage-provider/pull/349) fix: keep retrying when any blocksycner event handles failure


## v0.1.1

FEATURES
* [\#274](https://github.com/bnb-chain/greenfield-storage-provider/pull/274) feat: update stream record column names
* [\#275](https://github.com/bnb-chain/greenfield-storage-provider/pull/275) refactor: tasknode streaming process reduces memory usage
* [\#279](https://github.com/bnb-chain/greenfield-storage-provider/pull/283) feat: grpc client adds retry function
* [\#292](https://github.com/bnb-chain/greenfield-storage-provider/pull/292) feat: add table recreate func & block height metric for block sycner
* [\#295](https://github.com/bnb-chain/greenfield-storage-provider/pull/295) feat: support https protocol
* [\#296](https://github.com/bnb-chain/greenfield-storage-provider/pull/296) chore: change sqldb default config
* [\#299](https://github.com/bnb-chain/greenfield-storage-provider/pull/299) feat: add nat manager for p2p
* [\#304](https://github.com/bnb-chain/greenfield-storage-provider/pull/304) feat: support dns for p2p node
* [\#325](https://github.com/bnb-chain/greenfield-storage-provider/pull/325) feat: add universal endpoint
* [\#333](https://github.com/bnb-chain/greenfield-storage-provider/pull/333) fix: use EIP-4361 message template for off-chain-auth
* [\#339](https://github.com/bnb-chain/greenfield-storage-provider/pull/339) fix: permit anonymous users to access public object
* [\#347](https://github.com/bnb-chain/greenfield-storage-provider/pull/347) fix: add spdb and piece store metrics for downloader

BUGFIX
* [\#277](https://github.com/bnb-chain/greenfield-storage-provider/pull/277) fix: rcmgr leak for downloader service
* [\#278](https://github.com/bnb-chain/greenfield-storage-provider/pull/278) fix: uploader panic under db access error
* [\#279](https://github.com/bnb-chain/greenfield-storage-provider/pull/279) chore: change default rcmgr limit to no infinite
* [\#286](https://github.com/bnb-chain/greenfield-storage-provider/pull/286) fix: fix challenge memory is inaccurate
* [\#288](https://github.com/bnb-chain/greenfield-storage-provider/pull/288) fix: fix auth type v2 query object bug
* [\#306](https://github.com/bnb-chain/greenfield-storage-provider/pull/306) fix: fix multi update map bug and polish db error
* [\#337](https://github.com/bnb-chain/greenfield-storage-provider/pull/337) fix: permit anonymous users to access public objec


## v0.1.0

BUGFIX
* [\#258](https://github.com/bnb-chain/greenfield-storage-provider/pull/258) fix put object verify permission bug
* [\#264](https://github.com/bnb-chain/greenfield-storage-provider/pull/264) fix: fix payment apis nil pointer error
* [\#265](https://github.com/bnb-chain/greenfield-storage-provider/pull/265) fix: fix sa iam type to access s3
* [\#268](https://github.com/bnb-chain/greenfield-storage-provider/pull/268) feat: update buckets/objects order
* [\#270](https://github.com/bnb-chain/greenfield-storage-provider/pull/270) feat: update buckets/objects order
* [\#272](https://github.com/bnb-chain/greenfield-storage-provider/pull/272) fix: upgrade juno version for a property length fix

BUILD
* [\#259](https://github.com/bnb-chain/greenfield-storage-provider/pull/259) ci: fix release.yml uncorrect env var name
* [\#263](https://github.com/bnb-chain/greenfield-storage-provider/pull/263) feat: add e2e test to workflow


## v0.0.5

FEATURES
* [\#211](https://github.com/bnb-chain/greenfield-storage-provider/pull/211) feat: sp services add metrics
* [\#221](https://github.com/bnb-chain/greenfield-storage-provider/pull/221) feat: implement p2p protocol and rpc service
* [\#232](https://github.com/bnb-chain/greenfield-storage-provider/pull/232) chore: refine gRPC error code
* [\#235](https://github.com/bnb-chain/greenfield-storage-provider/pull/235) feat: implement metadata payment apis
* [\#244](https://github.com/bnb-chain/greenfield-storage-provider/pull/244) feat: update the juno version
* [\#246](https://github.com/bnb-chain/greenfield-storage-provider/pull/246) feat: resource manager

BUILD
* [\#231](https://github.com/bnb-chain/greenfield-storage-provider/pull/231) ci: add gosec checker


## v0.0.4

FEATURES
* [\#202](https://github.com/bnb-chain/greenfield-storage-provider/pull/202) feat: update get bucket apis
* [\#205](https://github.com/bnb-chain/greenfield-storage-provider/pull/205) fix: blocksyncer adapt event param to chain side and payment module added
* [\#206](https://github.com/bnb-chain/greenfield-storage-provider/pull/206) feat: support query quota and list read record
* [\#215](https://github.com/bnb-chain/greenfield-storage-provider/pull/215) fix: potential attack risks in on-chain storage module

IMPROVEMENT
* [\#188](https://github.com/bnb-chain/greenfield-storage-provider/pull/188) refactor: refactor metadata service
* [\#196](https://github.com/bnb-chain/greenfield-storage-provider/pull/196) docs: add sp docs
* [\#197](https://github.com/bnb-chain/greenfield-storage-provider/pull/197) refactor: rename stonenode, syncer to tasknode, recevier
* [\#200](https://github.com/bnb-chain/greenfield-storage-provider/pull/200) docs: refining readme
* [\#208](https://github.com/bnb-chain/greenfield-storage-provider/pull/208) docs: add block syncer config
* [\#209](https://github.com/bnb-chain/greenfield-storage-provider/pull/209) fix: block syncer db response style

BUGFIX
* [\#189](https://github.com/bnb-chain/greenfield-storage-provider/pull/189) fix: fix approval expired height bug
* [\#212](https://github.com/bnb-chain/greenfield-storage-provider/pull/212) fix: authv2 workflow
* [\#216](https://github.com/bnb-chain/greenfield-storage-provider/pull/216) fix: metadata buckets api

BUILD
* [\#179](https://github.com/bnb-chain/greenfield-storage-provider/pull/179) ci: add branch naming rules
* [\#198](https://github.com/bnb-chain/greenfield-storage-provider/pull/198) build: replace go1.19 with go1.18


## v0.0.3

FEATURES
* [\#169](https://github.com/bnb-chain/greenfield-storage-provider/pull/169) feat: piece store adds minio storage type
* [\#172](https://github.com/bnb-chain/greenfield-storage-provider/pull/172) feat: implement manager module
* [\#173](https://github.com/bnb-chain/greenfield-storage-provider/pull/173) feat: add check billing

IMPROVEMENT
* [\#154](https://github.com/bnb-chain/greenfield-storage-provider/pull/154) feat: syncer opt with chain data struct
* [\#156](https://github.com/bnb-chain/greenfield-storage-provider/pull/156) refactor: implement sp db, remove meta db and job db
* [\#157](https://github.com/bnb-chain/greenfield-storage-provider/pull/157) refactor: polish gateway module
* [\#162](https://github.com/bnb-chain/greenfield-storage-provider/pull/162) feat: add command for devops and config log
* [\#165](https://github.com/bnb-chain/greenfield-storage-provider/pull/165) feat: improve sync piece efficiency
* [\#171](https://github.com/bnb-chain/greenfield-storage-provider/pull/171) feat: add localup script


## v0.0.2

This release includes following features:
1. Implement the connection with the greenfield chain, and the upload and download of payload, including basic permission verification.
2. Implement the signer service for storage providers to sign the on-chain transactions.
3. Implement the communication of HTTP between SPs instead of gRPC.
* [\#131](https://github.com/bnb-chain/greenfield-storage-provider/pull/131) feat: add chain client to sp
* [\#119](https://github.com/bnb-chain/greenfield-storage-provider/pull/119) feat: implement signer service
* [\#128](https://github.com/bnb-chain/greenfield-storage-provider/pull/128) feat: stone node sends piece data to gateway
* [\#127](https://github.com/bnb-chain/greenfield-storage-provider/pull/127) feat: implement gateway challenge workflow
* [\#133](https://github.com/bnb-chain/greenfield-storage-provider/pull/133) fix: upgrade greenfield version to fix the signing bug
* [\#130](https://github.com/bnb-chain/greenfield-storage-provider/pull/130) fix: use env var to get bucket url



## v0.0.1

IMPROVEMENT
* [\#65](https://github.com/bnb-chain/greenfield-storage-provider/pull/65) feat: gateway add verify signature
* [\#43](https://github.com/bnb-chain/greenfield-storage-provider/pull/43) feat(uploader): add getAuth interface
* [\#68](https://github.com/bnb-chain/greenfield-storage-provider/pull/68) refactor: add jobdb v2 interface, objectID as primary key
* [\#70](https://github.com/bnb-chain/greenfield-storage-provider/pull/70) feat: change index from create object hash to object id
* [\#73](https://github.com/bnb-chain/greenfield-storage-provider/pull/73) feat(metadb): add sql metadb
* [\#82](https://github.com/bnb-chain/greenfield-storage-provider/pull/82) feat(stone_node): supports sending data to different storage provider
* [\#66](https://github.com/bnb-chain/greenfield-storage-provider/pull/66) fix: adjust the dispatching strategy of replica and inline data into storage provider
* [\#69](https://github.com/bnb-chain/greenfield-storage-provider/pull/69) fix: use multi-dimensional array to send piece data and piece hash
* [\#101](https://github.com/bnb-chain/greenfield-storage-provider/pull/101) fix: remove tokens from config and use env vars to load tokens
* [\#83](https://github.com/bnb-chain/greenfield-storage-provider/pull/83) chore(sql): polish sql workflow
* [\#87](https://github.com/bnb-chain/greenfield-storage-provider/pull/87) chore: add setup-test-env tool

BUILD
* [\#74](https://github.com/bnb-chain/greenfield-storage-provider/pull/74) ci: add docker release pipe
* [\#67](https://github.com/bnb-chain/greenfield-storage-provider/pull/67) ci: add commit lint, code lint and unit test ci files
* [\#85](https://github.com/bnb-chain/greenfield-storage-provider/pull/85) chore: add pull request template
* [\#105](https://github.com/bnb-chain/greenfield-storage-provider/pull/105) fix: add release action


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
3. Implement one-click deployment and one-click running test, which is convenient for developers and testers to experience the gnfd-sp.

* [\#7](https://github.com/bnb-chain/greenfield-storage-provider/pull/7) feat(gateway/uploader): add gateway and uploader skeleton
* [\#16](https://github.com/bnb-chain/greenfield-storage-provider/pull/16) Add secondary syncer service
* [\#17](https://github.com/bnb-chain/greenfield-storage-provider/pull/17) feat: implement of upload payload in stone hub side
* [\#29](https://github.com/bnb-chain/greenfield-storage-provider/pull/28) fix: ston node goroutine model
* [\#38](https://github.com/bnb-chain/greenfield-storage-provider/pull/38) feat: implement the challenge service
* [\#9](https://github.com/bnb-chain/greenfield-storage-provider/pull/9) add service lifecycle module
* [\#2](https://github.com/bnb-chain/greenfield-storage-provider/pull/2) add piecestore module
* [\#18](https://github.com/bnb-chain/greenfield-storage-provider/pull/18) feat: add job meta orm
* [\#60](https://github.com/bnb-chain/greenfield-storage-provider/pull/60) test: add run cases
