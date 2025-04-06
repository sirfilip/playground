package task_manager

func Fmap[T, U any](src []T, mapper func(T) (U, error)) ([]U, error) {
	mapped := make([]U, 0, len(src))
	for _, t := range src {
		u, err := mapper(t)
		if err != nil {
			return nil, err
		}
		mapped = append(mapped, u)
	}
	return mapped, nil
}

func MustFmap[T, U any](src []T, mapper func(T) (U, error)) []U {
	mapped := make([]U, 0, len(src))
	for _, t := range src {
		u, err := mapper(t)
		if err != nil {
			panic(err)
		}
		mapped = append(mapped, u)
	}
	return mapped
}
