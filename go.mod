module github.com/joelanford/declcfg

go 1.16

require (
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/blang/semver/v4 v4.0.0
	github.com/bshuster-repo/logrus-logstash-hook v1.0.2 // indirect
	github.com/jinzhu/copier v0.3.2
	github.com/operator-framework/operator-registry v0.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	k8s.io/apimachinery v0.20.6
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace github.com/operator-framework/operator-registry => github.com/joelanford/operator-registry v1.12.2-0.20210719211906-d36dab96a303
