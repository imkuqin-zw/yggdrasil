package xgorm

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// PaginationParam 分页查询条件
type PaginationParam struct {
	Pagination bool   // 是否使用分页查询
	OnlyCount  bool   // 是否仅查询count
	NoCount    bool   // 不需要进行count
	Current    uint32 // 当前页
	PageSize   uint32 // 页大小
}

// GetCurrent 获取当前页
func (a PaginationParam) GetCurrent() uint32 {
	return a.Current
}

// GetPageSize 获取页大小
func (a PaginationParam) GetPageSize() uint32 {
	pageSize := a.PageSize
	if a.PageSize == 0 {
		pageSize = 100
	}
	return pageSize
}

type PaginationResult struct {
	Total    uint32
	Current  uint32
	PageSize uint32
}

// WrapPageQuery 包装带有分页的查询
func WrapPageQuery(db *gorm.DB, pp PaginationParam, out interface{}) (*PaginationResult, error) {
	if pp.OnlyCount {
		var count int64
		table, _ := rowSliceElement(out)
		err := db.Model(table).Count(&count).Error
		if err != nil {
			return nil, err
		}
		return &PaginationResult{Total: uint32(count)}, nil
	} else if !pp.Pagination {
		err := db.Find(out).Error
		return nil, err
	}

	total, err := findPage(db, pp, out)
	if err != nil {
		return nil, err
	}

	return &PaginationResult{
		Total:    uint32(total),
		Current:  pp.GetCurrent(),
		PageSize: pp.GetPageSize(),
	}, nil
}

// FindPage 查询分页数据
func findPage(db *gorm.DB, pp PaginationParam, out interface{}) (int64, error) {
	var count int64
	if !pp.NoCount {
		table, _ := rowSliceElement(out)
		err := db.Model(table).Count(&count).Error
		if err != nil {
			return 0, err
		} else if count == 0 {
			return count, nil
		}
	}

	current, pageSize := pp.GetCurrent(), pp.GetPageSize()
	if current > 0 && pageSize > 0 {
		db = db.Offset((int(current) - 1) * int(pageSize)).Limit(int(pageSize))
	} else if pageSize > 0 {
		db = db.Limit(int(pageSize))
	}

	err := db.Find(out).Error
	return count, err
}

// FindOne 查询单条数据
func FindOne(db *gorm.DB, out interface{}) (bool, error) {
	result := db.First(out)
	if err := result.Error; err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func rowSliceElement(rowsSlicePtr interface{}) (interface{}, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Map {
		return 0, errors.New("needs a pointer to a slice or a map")
	}

	sliceElementType := sliceValue.Type().Elem()
	if sliceElementType.Kind() == reflect.Ptr {
		sliceElementType = sliceElementType.Elem()
	}
	return reflect.New(sliceElementType).Interface(), nil
}
