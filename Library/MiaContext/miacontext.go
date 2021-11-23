package MiaContext 
import(
	"context"
	"time"
)
func TimeoutContext(timeout int) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	go func() {
		time.Sleep(time.Second * time.Duration(timeout))
		cancel()
	}()
	return ctx
}