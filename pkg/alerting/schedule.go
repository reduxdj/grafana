package alerting

import (
	"bytes"
	"text/template"

	m "github.com/grafana/grafana/pkg/models"
)

type Schedule struct {
	//Id           int64
	//OrgId        int64
	//DataSourceId int64
	Freq       uint32
	Offset     uint8 // offset on top of "even" minute/10s/.. intervals
	Definition CheckDef
}

func getSchedules() ([]Schedule, error) {
	// this may not be the ideal query to run, depending on where we store scheduling info,
	// for now let's pretend we use this.
	// we also need endpoint slug, and slugs for all the collectors
	// see https://github.com/raintank/grafana/issues/83
	// TODO: for some reason sometimes a handler is, then isn't, then is again, found for this :?
	/*
		query := m.GetMonitorsQuery{}

		fmt.Println(">getSchedules() dispatching GetMonitorsQuery")
		if err := bus.Dispatch(&query); err != nil {
			fmt.Println(">getSchedules() got error!", err)
			return nil, err
		}

		fmt.Println(">getSchedules() got result!")
		spew.Dump(query.Result)
	*/
	schedules := make([]Schedule, 0)
	fakeRes := []*m.MonitorDTO{
		&m.MonitorDTO{
			HealthSettings: m.MonitorHealthSettingDTO{
				NumCollectors: 2,
				Steps:         2,
			},
		},
	}

	for _, monitor := range fakeRes { //query.Result {
		schedules = append(schedules, buildScheduleForMonitor(monitor))
	}

	return schedules, nil
}

func buildScheduleForMonitor(monitor *m.MonitorDTO) Schedule {
	type TempJoinedStruct struct {
		Endpoint       string
		Collector      string
		Type           string
		State          string // warn, error
		HealthSettings m.MonitorHealthSettingDTO
	}

	temp := TempJoinedStruct{
		Endpoint:       "dieter_plaetinck_be", // as in graphite metric
		Collector:      "*",
		Type:           "http", // as in graphite metric: ping, http, https, dns. does this come from MonitorTypeId ?
		State:          "warn", // as in.. ok, warn, error
		HealthSettings: monitor.HealthSettings,
	}
	// TODO we must know the interval . is that what Frequency is?
	tpl := `sum(graphite("{{.Endpoint}}.{{.Collector}}.network.{{.Type}}.{{.State}}_state", "{{.HealthSettings.Steps}}m", "", "") > 0) > {{.HealthSettings.NumCollectors}}`
	var t = template.Must(template.New("query").Parse(tpl))
	var b bytes.Buffer
	err := t.Execute(&b, temp)
	if err != nil {
		panic(err)
	}
	s := Schedule{
		Freq:   60, // TODO discuss how to do this with anthony
		Offset: 5,
		Definition: CheckDef{
			CritExpr: b.String(),
			WarnExpr: "0", // for now we have only good or bad. so only crit is needed
		},
	}
	return s
}