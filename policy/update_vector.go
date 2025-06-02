package policy

import (
	"math"
	"time"
)

type UserVector struct {
	Embedding []float32
	WeightSum float64
	UpdatedAt time.Time
	tau       float64
}

// NewUserVector 傳入舊事件向量、權重半衰期、舊權重總和、更新時間
//
// 如果之前沒有資料則 embedding 傳入一個 make 好的陣列（長度為 1536），weightSum 傳入 0，updatedAt 傳入 time.Now()
func NewUserVector(embedding []float32, halfLifeDays, weightSum float64, updatedAt time.Time) *UserVector {
	v := &UserVector{
		Embedding: embedding,
		WeightSum: weightSum,
		UpdatedAt: updatedAt,
	}
	v.tau = halfLifeDays * 24 * 3600 / math.Log(2)

	return v
}

// UpdateUserVec 更新使用者向量資料，傳入新事件向量、此事件的基本權重
func (v *UserVector) UpdateUserVec(eventVec []float32, baseWeight float64) {
	dt := time.Since(v.UpdatedAt).Seconds()
	if dt > 1 {
		decay := math.Exp(-dt / v.tau)
		for i := range v.Embedding {
			v.Embedding[i] *= float32(decay)
		}
		v.WeightSum += decay
	}

	// 做增量平均
	for i := range eventVec {
		v.Embedding[i] = (v.Embedding[i]*float32(v.WeightSum) + float32(baseWeight)*eventVec[i]) /
			float32(v.WeightSum+baseWeight)
	}

	v.WeightSum += baseWeight
	v.UpdatedAt = time.Now()
}

// GetVectorAt 獲取某個時間節點對應的向量，透過及時計算半衰期回傳受時間影響後的向量值
func (v *UserVector) GetVectorAt(at time.Time) []float32 {
	dt := at.Sub(v.UpdatedAt).Seconds()
	// 若要求時間早於更新時間，則直接回傳目前 embed 向量
	if dt <= 0 {
		return v.Embedding
	}

	decay := math.Exp(-dt / v.tau)
	vct := make([]float32, len(v.Embedding))
	for i := range vct {
		vct[i] = v.Embedding[i] * float32(decay)
	}

	return vct
}
