package schedule

type ScheduleProvider interface {
	GetSchedule(teacherID string) (ScheduleResponse, error)
}
