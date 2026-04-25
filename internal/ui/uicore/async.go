package uicore

// import (
// 	"sync/atomic"

// 	"github.com/crispyarty/LinkInterceptor/internal/ui/dispatch"
// )

// type SyncResult[D any] struct {
// 	Items []D
// 	Cache map[string]D
// }

// func ProcessAsync[D any, U any](
// 	version *atomic.Int64,
// 	data D,
// 	transform func(D) U,
// 	apply func(U),
// ) {
// 	// 1. Increment version (Thread-safe)
// 	v := version.Add(1)

// 	go func() {
// 		// 2. Background: Run the transformation
// 		result := transform(data)

// 		// 3. Dispatch: Back to main thread
// 		dispatch.Actions <- func() {
// 			// 4. Version Check: Only apply if no newer task started
// 			if v == version.Load() {
// 				apply(result)
// 			}
// 		}
// 	}()
// }

// Example:
// func (s *SettingsUi) HandleData(data *system.Settings) {
//     uicore.ProcessAsync(&s.version, data, s.transform, s.apply)
// }
