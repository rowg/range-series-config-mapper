package config_interval

import "time"

type ConfigInterval struct {
	Start  time.Time
	End    time.Time
	Config string
}

func (timeInt ConfigInterval) ContainsTime(timestamp time.Time) bool {
	return (timestamp.After(timeInt.Start) || timestamp.Equal(timeInt.Start)) && timestamp.Before(timeInt.End)
}
