package entity

import "time"

func Set[D, S any](dst *D, src S, proc func(S) (D, error)) error {
	val, err := proc(src)
	if err != nil {
		return err
	}

	*dst = val
	return nil
}

func SetWithUpdate[D, S any](dst *D, src S, proc func(S) (D, error), update *time.Time) error {
	if err := Set(dst, src, proc); err != nil {
		return err
	}

	if update != nil {
		*update = time.Now()
	}

	return nil
}
