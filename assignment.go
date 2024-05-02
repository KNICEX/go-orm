package orm

type Assignment struct {
	col string
	val any
}

func (a Assignment) assign() {}

// 实现 SetAble 标记接口， Assignment 可以作为一个SET语句的列
func (a Assignment) setAble() {}

func Assign(col string, val any) Assignment {
	return Assignment{
		col: col,
		val: val,
	}
}
