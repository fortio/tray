package ray

// AABB represents an axis-aligned bounding box in 3D space.
type AABB [3]Interval

func NewAABB(a, b Vec3) AABB {
	return AABB{
		OrderedInterval(a.x, b.x),
		OrderedInterval(a.y, b.y),
		OrderedInterval(a.z, b.z),
	}
}

func UnionAABB(box1, box2 AABB) AABB {
	return AABB{
		Union(box1[0], box2[0]),
		Union(box1[1], box2[1]),
		Union(box1[2], box2[2]),
	}
}

//nolint:nestif // yeah.
func (box AABB) Hit(ray *Ray, rayT Interval) bool {
	ro := ray.Origin.Components()
	rd := ray.Direction.Components()
	for a := range 3 {
		i := box[a]
		adinv := 1.0 / rd[a]
		t0 := (i.Start - ro[a]) * adinv
		t1 := (i.End - ro[a]) * adinv
		if t0 < t1 {
			if t0 > rayT.Start {
				rayT.Start = t0
			}
			if t1 < rayT.End {
				rayT.End = t1
			}
		} else {
			if t1 > rayT.Start {
				rayT.Start = t1
			}
			if t0 < rayT.End {
				rayT.End = t0
			}
		}
		if rayT.End <= rayT.Start {
			return false
		}
	}
	return true
}
