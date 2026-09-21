package silk

// nsqQuant4 batches independent delayed-decision states. Residuals are already
// dithered and clamped to [-31<<10, 30<<10], as in silk/NSQ_del_dec.c.
// Offset is one of the codec quantization offsets (32, 100, 240).
type nsqQuant4 struct {
	r, rd                [4]int32
	q0, q1, cost0, cost1 [4]int32
}

func quantize4Go(batch *nsqQuant4, offsetQ10i32, lambdaQ10i32 int32) {
	useRDO := lambdaQ10i32 > 2048
	rdoOffset := lambdaQ10i32/2 - 512
	for k := range 4 {
		rQ10 := batch.r[k]
		q1Q10 := rQ10 - offsetQ10i32
		q1Q0 := q1Q10 >> 10
		if useRDO {
			if q1Q10 > rdoOffset {
				q1Q0 = (q1Q10 - rdoOffset) >> 10
			} else if q1Q10 < -rdoOffset {
				q1Q0 = (q1Q10 + rdoOffset) >> 10
			} else if q1Q10 < 0 {
				q1Q0 = -1
			} else {
				q1Q0 = 0
			}
		}
		var q2Q10, rd1Q10, rd2Q10 int32
		if q1Q0 > 0 {
			q1Q10 = (q1Q0 << 10) - quantLevelAdjQ10 + offsetQ10i32
			q2Q10 = q1Q10 + 1024
			rd1Q10 = silk_SMULBB(q1Q10, lambdaQ10i32)
			rd2Q10 = silk_SMULBB(q2Q10, lambdaQ10i32)
		} else if q1Q0 == 0 {
			q1Q10 = offsetQ10i32
			q2Q10 = q1Q10 + 1024 - quantLevelAdjQ10
			rd1Q10 = silk_SMULBB(q1Q10, lambdaQ10i32)
			rd2Q10 = silk_SMULBB(q2Q10, lambdaQ10i32)
		} else if q1Q0 == -1 {
			q2Q10 = offsetQ10i32
			q1Q10 = q2Q10 - 1024 + quantLevelAdjQ10
			rd1Q10 = silk_SMULBB(-q1Q10, lambdaQ10i32)
			rd2Q10 = silk_SMULBB(q2Q10, lambdaQ10i32)
		} else {
			q1Q10 = (q1Q0 << 10) + quantLevelAdjQ10 + offsetQ10i32
			q2Q10 = q1Q10 + 1024
			rd1Q10 = silk_SMULBB(-q1Q10, lambdaQ10i32)
			rd2Q10 = silk_SMULBB(-q2Q10, lambdaQ10i32)
		}
		rrQ10 := rQ10 - q1Q10
		rd1Q10 = silk_SMLABB(rd1Q10, rrQ10, rrQ10) >> 10
		rrQ10 = rQ10 - q2Q10
		rd2Q10 = silk_SMLABB(rd2Q10, rrQ10, rrQ10) >> 10

		if rd1Q10 < rd2Q10 {
			batch.cost0[k] = batch.rd[k] + rd1Q10
			batch.cost1[k] = batch.rd[k] + rd2Q10
			batch.q0[k] = q1Q10
			batch.q1[k] = q2Q10
		} else {
			batch.cost0[k] = batch.rd[k] + rd2Q10
			batch.cost1[k] = batch.rd[k] + rd1Q10
			batch.q0[k] = q2Q10
			batch.q1[k] = q1Q10
		}
	}
}
