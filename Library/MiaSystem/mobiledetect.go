package MiaSystem
import (
		"net/http"
)
func IsMobile(r *http.Request) bool {
	detect := NewMobileDetect(r, nil)
	if detect.IsMobile() || detect.IsTablet(){
		if detect.IsMobile() && detect.IsTablet(){
	//		fmt.Println("Hello, this is Tablet")
		}else {
	//		fmt.Println("Hello, this is Mobile")
		}
		return true
	}else {
	//	fmt.Println("Hello, this is Desktop")
		return false
	}
}