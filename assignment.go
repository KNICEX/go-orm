package orm

type Assignment struct {
	name string
	val  any
}

func (a Assignment) assign() {}

// 实现 SetAble 标记接口， Assignment 可以作为一个SET语句的列
func (a Assignment) setAble() {}

func Assign(col string, val any) Assignment {
	return Assignment{
		name: col,
		val:  val,
	}
}
