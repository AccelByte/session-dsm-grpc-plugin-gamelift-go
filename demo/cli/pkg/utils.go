// Copyright (c) 2024 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package sessiondsmdemo

import (
	"math/rand"
	"time"
)

func Ptr[K comparable](value K) *K {
	return &value
}

func Val[K any](ptr *K) K {
	if ptr != nil {
		return *ptr
	}
	var empty K

	return empty
}

func RandomString(characters string, length int) string {
	result := make([]byte, length)
	rand.Seed(time.Now().Unix())
	for i := 0; i < length; i++ {
		for {
			result[i] = characters[rand.Intn(len(characters))]
			if i > 0 && result[i-1] == result[i] {
				continue
			}

			break
		}
	}

	return string(result)
}
