package config

import (
	"os"

	"github.com/BurntSushi/toml"

	"tianxian.com/tee-signer-core/log"
)

func LoadConfig(configPath string) {
	Conf = new(Config)
	_, err := toml.DecodeFile(configPath, Conf)
	if err != nil {
		log.Entry.Fatal(err)
	}

	Conf.EmptyInterval = max(Conf.EmptyInterval, 100)
	Conf.ErrInterval = max(Conf.ErrInterval, 5_000)

	if len(Conf.Bean) == 0 {
		log.Entry.Fatal("no beans")
	}
	if len(Conf.Bean) == 1 {
		log.AddField("chain", Conf.Bean[0].Chain)
	}

	log.AddField("component", "pump")
	var hostname string
	if hostname, err = os.Hostname(); err != nil {
		hostname = "localhost"
	}
	log.AddField("host", hostname)

	if Conf.Log.Stdout.Enable {
		log.AddConsoleOut(Conf.Log.Stdout.Level)
	} else {
		log.Entry.Warn("no standard output log now")
		log.DisableDefaultConsole()
	}
	err = log.AddFileOut(Conf.Log.File.Path, 5, Conf.Log.File.MaxAge)
	if err != nil {
		log.Entry.Fatal(err)
	}
	if Conf.Log.Kafka.Enable {
		err = log.AddKafkaHook(Conf.Log.Kafka.Topic, Conf.Log.Kafka.Brokers, Conf.Log.Kafka.Level)
		if err != nil {
			log.Entry.Fatal(err)
		}
	}
}

var Conf *Config

type Config struct {
	// pump 处理完成上述任务后, 没有新任务时的休眠时间. 不同类型任务之间无影响. 最小值 100
	EmptyInterval int64 `toml:"empty_interval_ms"`
	// pump 处理上述报错后的休眠时间. 不同类型任务之间无影响. 最小值 5_000
	ErrInterval int64 `toml:"err_interval_ms"`
	Log      struct {
		Stdout stdout   `toml:"stdout"`
		File   file     `toml:"file"`
		Kafka  kafkalog `toml:"kafka"`
	} `toml:"log"`
}

type stdout struct {
	Enable bool `toml:"enable"`
	Level  int  `toml:"level"`
}

type file struct {
	Path   string `toml:"path"`
	MaxAge int    `toml:"max_age"`
}

type kafkalog struct {
	Enable  bool     `toml:"enable"`
	Level   int      `toml:"level"`
	Topic   string   `toml:"topic"`
	Brokers []string `toml:"brokers"`
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
