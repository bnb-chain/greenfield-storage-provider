module github.com/bnb-chain/greenfield-storage-provider

go 1.19

replace (
	// TODO: point to develop branch, will be changed to v0.0.6 after greenfield-sdk-go v0.0.6 released
	github.com/bnb-chain/greenfield => github.com/bnb-chain/greenfield v0.0.0-20230220084048-b09d0197ee41
	github.com/cosmos/cosmos-sdk => github.com/bnb-chain/gnfd-cosmos-sdk v0.0.2-0.20230216105555-b94a14479b11
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/tendermint/tendermint => github.com/bnb-chain/gnfd-tendermint v0.0.1
)

require (
	cosmossdk.io/errors v1.0.0-beta.7
	github.com/GeertJohan/go.linenoise v0.0.0-20141120151038-1918ff89d613
	github.com/aws/aws-sdk-go v1.44.159
	github.com/bnb-chain/greenfield v0.0.5
	github.com/bnb-chain/greenfield-common/go v0.0.0-20230222034925-1fa3ff1449c3
	github.com/bnb-chain/greenfield-go-sdk v0.0.0-20230222033526-17e6b63b80bc
	github.com/bytedance/gopkg v0.0.0-20221122125632-68358b8ecec6
	github.com/cloudflare/cfssl v1.6.3
	github.com/cosmos/cosmos-proto v1.0.0-beta.1
	github.com/cosmos/cosmos-sdk v0.46.4
	github.com/ethereum/go-ethereum v1.10.19
	github.com/golang/mock v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/reedsolomon v1.11.7
	github.com/looplab/fsm v1.0.0
	github.com/naoina/toml v0.1.2-0.20170918210437-9fafd6967416
	github.com/oleiade/lane v1.0.1
	github.com/pkg/sftp v1.13.5
	github.com/stretchr/testify v1.8.1
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/urfave/cli v1.22.10
	github.com/viki-org/dnscache v0.0.0-20130720023526-c70c1f23c5d8
	go.uber.org/multierr v1.9.0
	go.uber.org/zap v1.24.0
	golang.org/x/exp v0.0.0-20230111222715-75897c7a292a
	google.golang.org/grpc v1.52.3
	google.golang.org/protobuf v1.28.2-0.20220831092852-f930b1dc76e8
	gorm.io/driver/mysql v1.4.4
	gorm.io/gorm v1.24.2
)

require (
	cosmossdk.io/math v1.0.0-beta.3 // indirect
	filippo.io/edwards25519 v1.0.0-rc.1 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.1 // indirect
	github.com/ChainSafe/go-schnorrkel v0.0.0-20200405005733-88cbf1b4c40d // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/btcsuite/btcd v0.22.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/confio/ics23/go v0.7.0 // indirect
	github.com/cosmos/btcutil v1.0.4 // indirect
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/cosmos/gogoproto v1.4.4 // indirect
	github.com/cosmos/gorocksdb v1.2.0 // indirect
	github.com/cosmos/iavl v0.19.4 // indirect
	github.com/cosmos/ledger-cosmos-go v0.11.1 // indirect
	github.com/cosmos/ledger-go v0.9.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dustin/go-humanize v1.0.1-0.20200219035652-afde56e7acac // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/evmos/ethermint v0.6.1-0.20220919141022-34226aa7b1fa // indirect
	github.com/ferranbt/fastssz v0.0.0-20210905181407-59cf6761a7d5 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-kit/kit v0.12.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/protobuf v1.3.3 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/gtank/merlin v0.1.1 // indirect
	github.com/gtank/ristretto255 v0.1.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hdevalence/ed25519consensus v0.0.0-20220222234857-c00d1f31bab3 // indirect
	github.com/herumi/bls-eth-go-binary v0.0.0-20210917013441-d37c07cfda4e // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.3 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20210601165009-122bf33a46e0 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/prysmaticlabs/eth2-types v0.0.0-20210303084904-c9735a06829d // indirect
	github.com/prysmaticlabs/prysm v0.0.0-20220124113610-e26cde5e091b // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/regen-network/cosmos-proto v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/rs/zerolog v1.29.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/cobra v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.13.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/supranational/blst v0.3.8-0.20220526154634-513d2456b344 // indirect
	github.com/tendermint/btcd v0.1.1 // indirect
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15 // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	github.com/tendermint/tendermint v0.34.22 // indirect
	github.com/tendermint/tm-db v0.6.7 // indirect
	github.com/thomaso-mirodin/intmath v0.0.0-20160323211736-5dc6d854e46e // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/wealdtech/go-bytesutil v1.1.1 // indirect
	github.com/wealdtech/go-eth2-types/v2 v2.5.2 // indirect
	github.com/wealdtech/go-eth2-util v1.6.3 // indirect
	github.com/zondax/hid v0.9.1-0.20220302062450-5552068d2266 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20221207170731-23e4bf6bdc37 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
