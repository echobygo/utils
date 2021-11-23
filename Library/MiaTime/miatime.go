package MiaTime
import "time"
func   GetMaxTime(time1,time2 time.Time) time.Time{
	if time1.After(time2){
		return time1
	}
	return time2
}