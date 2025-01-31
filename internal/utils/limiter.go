package utils

const requestLimit = 5 // Максимум 5 запросов одновременно
var requestSemaphore = make(chan struct{}, requestLimit)

func RerformRequest(fn func() ([]byte, error)) ([]byte, error) {
	requestSemaphore <- struct{}{}        // Захват места в семафоре
	defer func() { <-requestSemaphore }() // Освобождение после выполнения

	data, err := fn()
	if err != nil {
		return nil, err
	}
	return data, nil
}
