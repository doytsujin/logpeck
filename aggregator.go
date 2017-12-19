package logpeck

import (
	"github.com/Sirupsen/logrus"
	"sort"
	"strconv"
)

type AggregatorConfig struct {
	Tags         []string `json:"tags"`
	Aggregations []string `json:"aggregations"`
	Target       string   `json:"target"`
	Time         string   `json:"time"`
}

type Aggregator struct {
	interval          int64
	name              string
	aggregatorConfigs map[string]AggregatorConfig
	buckets           map[string]map[string][]int
	postTime          int64
}

func NewAggregator(interval int64, name string, aggregators *map[string]AggregatorConfig) *Aggregator {
	aggregator := &Aggregator{
		interval:          interval,
		name:              name,
		aggregatorConfigs: *aggregators,
		buckets:           make(map[string]map[string][]int),
		postTime:          0,
	}
	return aggregator
}

func getSampleTime(ts int64, interval int64) int64 {
	return ts / interval
}

func (p *Aggregator) IsDeadline(timestamp int64) bool {
	interval := p.interval
	nowTime := getSampleTime(timestamp, interval)
	if p.postTime != nowTime {
		return true
	}
	return false
}

func (p *Aggregator) Record(fields map[string]interface{}) int64 {
	//get sender
	//influxDbConfig := p.Config.SenderConfig.Config.(InfluxDbConfig)
	bucketName := fields[p.name].(string)
	bucketTag := ""
	aggregatorConfig := p.aggregatorConfigs[bucketName]
	tags := aggregatorConfig.Tags
	aggregations := aggregatorConfig.Aggregations
	target := aggregatorConfig.Target
	time := aggregatorConfig.Time
	for i := 0; i < len(tags); i++ {
		bucketTag += "," + tags[i] + "=" + fields[tags[i]].(string)
	}
	int_bool := false
	for i := 0; i < len(aggregations); i++ {
		if aggregations[i] != "cnt" {
			int_bool = true
		}
	}
	aggValue := fields[target].(string)

	//get time
	now, err := strconv.ParseInt(fields[time].(string), 10, 64)
	if err != nil {
		logrus.Debug("[Record] timestamp:%v can't use strconv.ParseInt", fields[time].(string))
		return now
	}

	if _, ok := p.buckets[bucketName]; !ok {
		p.buckets[bucketName] = make(map[string][]int)
	}
	if int_bool == false {
		p.buckets[bucketName][bucketTag] = append(p.buckets[bucketName][bucketTag], 1)
	} else {
		aggValue, err := strconv.Atoi(aggValue)
		if err != nil {
			logrus.Debug("[Record] target:%v can't use strconv.Atoi", aggValue)
			return now
		}
		p.buckets[bucketName][bucketTag] = append(p.buckets[bucketName][bucketTag], aggValue)
	}
	return now
}

func getAggregation(targetValue []int, aggregations []string) map[string]int {
	aggregationResults := map[string]int{}
	cnt := len(targetValue)
	avg := 0
	sum := 0
	sort.Ints(targetValue)
	for _, value := range targetValue {
		sum += value
	}
	avg = sum / cnt
	for i := 0; i < len(aggregations); i++ {
		switch aggregations[i] {
		case "cnt":
			aggregationResults["cnt"] = len(targetValue)
		case "avg":
			aggregationResults["avg"] = avg
		default:
			if aggregations[i][0] == 'p' {
				proportion, err := strconv.Atoi(aggregations[i][1:])
				if err != nil {
					panic(aggregations[i])
				}
				percentile := targetValue[cnt*proportion/100-1]
				aggregationResults[aggregations[i]] = percentile
			}
		}
	}
	return aggregationResults
}

func (p *Aggregator) Dump(timestamp int64) map[string]interface{} {
	fields := map[string]interface{}{}
	//now := strconv.FormatInt(timestamp, 10)
	for bucketName, bucketTag_value := range p.buckets {
		for bucketTag, targetValue := range bucketTag_value {
			aggregations := p.aggregatorConfigs[bucketName].Aggregations
			fields[bucketName+bucketTag] = getAggregation(targetValue, aggregations)
		}
	}
	fields["timestamp"] = timestamp
	p.postTime = getSampleTime(timestamp, p.interval)
	p.buckets = map[string]map[string][]int{}
	return fields
}
