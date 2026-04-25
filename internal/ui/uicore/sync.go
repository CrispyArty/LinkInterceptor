package uicore

type Syncable[T any] interface {
	Update(dataItem T)
}

// Syncs list of items for new data
// D - Data item
// U - Ui item(must implement Syncable)
// K - Comparable value that will act as a key
func SyncList[D any, U Syncable[D], K comparable](
	data []D,
	oldCache map[K]U,
	getKey func(D) K,
	factory func() U,
) ([]U, map[K]U) {
	newItems := make([]U, len(data))
	newCache := make(map[K]U, len(data))

	for id, item := range data {
		key := getKey(item)

		if old, exists := oldCache[key]; exists {
			old.Update(item)
			newItems[id] = old
		} else {
			newItem := factory()
			newItem.Update(item)
			newItems[id] = newItem
		}

		newCache[key] = newItems[id]
	}

	return newItems, newCache
}
