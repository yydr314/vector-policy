# 使用者喜好向量計算說明

## 1. 目的

建立一套**即時、遞減權重**的使用者偏好向量，讓近期且重要的互動（如「參加活動」）在推薦演算法中影響力更大，而久遠或次要的互動影響力逐步衰減。

## 2. 重要符號

| 符號          | 定義                           |
|-------------|------------------------------|
| $\mathbf P$ | 使用者偏好向量（長度 1536）             |
| $W$         | 權重總和（WeightSum）              |
| $\mathbf v$ | 新事件的 embedding 向量（同樣 1536 維） |
| $w$         | 新事件的基礎權重（依 event type 決定）    |
| $\tau$      | 欲設定的權重**平均壽命**（由半衰期計算）       |
| $\Delta t$  | 與上一次更新的時間間隔（秒）               |

## 3. 時間衰減模型

對舊資料整體施加指數衰減係數

$$
r = e^{-\Delta t / \tau}
$$

半衰期 $T_{1/2}$ 轉換至 $\tau$ 的公式：

$$
\tau = \frac{T_{1/2}}{\ln 2}
$$

> 程式碼中：`tau = halfLifeDays * 24 * 3600 / math.Log(2)`

## 4. 增量更新公式

1. **先衰減既有值**

   $$\mathbf P_d = r\,\mathbf P_{old}, \quad W_d = r\,W_{old}$$
2. **加入新事件**（已含基礎權重）

   $$\mathbf P_{new} = \frac{\mathbf P_d\,W_d + w\,\mathbf v}{W_d + w}$$

   $$W_{new} = W_d + w$$

上述運算複雜度 $\Theta(\text{向量維度})$，不必重新掃描歷史。

## 5. 程式碼對照

### 5.1 `NewUserVector`

* 建立結構，並以半衰期換算 `tau`。

### 5.2 `UpdateUserVec`

1. 以 `time.Since(UpdatedAt)` 計算 $\Delta t$；若間隔大於 1 秒則執行衰減：

   ```go
   decay := math.Exp(-dt / v.tau)
   v.Embedding[i] *= float32(decay)
   v.WeightSum *= decay
   ```
2. 使用增量公式更新每一維：

   ```go
   v.Embedding[i] = (v.Embedding[i]*float32(v.WeightSum) + float32(baseWeight)*eventVec[i]) / float32(v.WeightSum + baseWeight)
   ```
3. 累加 `WeightSum` 並刷新 `UpdatedAt`。

### 5.3 `GetVectorAt`

* 依查詢時間 `at` 再做一次衰減，保證回傳的向量代表「當下」的權重效果。

## 6. 權重設定範例

| 事件類型         | 基礎權重範例 $w$ |
|--------------|------------|
| 參加活動(`join`) | 3.0        |
| 查看活動(`view`) | 1.0        |
| 搜尋(`search`) | 0.5        |

> 權重與半衰期可依業務需求調整。
