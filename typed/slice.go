package typed

import (
	"errors"
	"idie/util"
	"reflect"
)

type sliceCompareFunc func(item interface{}, nLoopItem interface{}) bool

type Slice struct {
	//public
	Items []interface{}

	//private
	lastStartOffset int
	lastEndOffset   int
	lastPage        int
}

// indexOf is a method of Slice to get index of an item from Items
func (this *Slice) IndexOf(item interface{}, cmpFunc sliceCompareFunc) int {
	for i, v := range this.Items {
		if cmpFunc != nil {
			if cmpFunc(item, v) {
				return i
			}
			continue
		}

		if v == item {
			return i
		}
	}
	return -1
}

// append is a method of Slice to add an item to end of Items
func (this *Slice) Append(item interface{}) {
	this.Items = append(this.Items, item)
}

// prepend is a method of Slice to add an item to start of Items
func (this *Slice) Prepend(item interface{}) {
	this.Items = append([]interface{}{item}, this.Items...)
}

// remove is a method of Slice to remove an item from Items
func (this *Slice) Remove(item interface{}, cmpFunc sliceCompareFunc) {
	indexOfItem := this.IndexOf(item, cmpFunc)
	if indexOfItem != -1 {
		this.Items = append(this.Items[:indexOfItem], this.Items[indexOfItem+1:]...)
	}
}

// removeAt is a method of Slice to remove an item at index from Items
func (this *Slice) RemoveAt(index int) (interface{}, error) {
	if index < 0 || index >= this.Count() {
		return nil, errors.New("index out of range")
	}

	item := this.Items[index]
	this.Items = append(this.Items[:index], this.Items[index+1:]...)
	return item, nil
}

func (this *Slice) IsItemAtStructAndHasField(index int, field string) bool {
	if index < 0 || index >= this.Count() {
		return false
	}

	item, _ := this.Get(index)
	if item == nil {
		return false
	}

	ok, structVal := util.IsStruct(item)
	if !ok {
		return false
	}

	fieldVal := structVal.FieldByName(field)
	if fieldVal == (reflect.Value{}) {
		return false
	}

	return true
}

func (this *Slice) IndexOfByStructFieldAndValue(field string, value interface{}) int {
	for i := range this.Items {
		if !this.IsItemAtStructAndHasField(i, field) {
			continue
		}

		item, _ := this.Get(i)
		_, structVal := util.IsStruct(item)
		fieldVal := structVal.FieldByName(field)
		if fieldVal.Interface() == value {
			return i
		}
	}

	return -1
}

func (this *Slice) GetItemByStructFieldAndValue(field string, value interface{}) interface{} {
	index := this.IndexOfByStructFieldAndValue(field, value)
	if index == -1 {
		return nil
	}

	item, _ := this.Get(index)
	return item
}

// contains is a method of Slice to check if Items contains an item
func (this *Slice) Contains(item interface{}, cmpFunc sliceCompareFunc) bool {
	return this.IndexOf(item, cmpFunc) != -1
}

func (this *Slice) CountContains(item interface{}, cmpFunc sliceCompareFunc) int {
	var count int
	for _, nLoopItem := range this.Items {
		if cmpFunc != nil {
			if cmpFunc(item, nLoopItem) {
				count++
			}
			continue
		}

		if nLoopItem == item {
			count++
		}
	}
	return count
}

// unique is a method of Slice to remove duplicate items from Items
// . this might work for Numeric types and String type only
func (this *Slice) Unique() {
	uniqueIndexes := make(map[interface{}]bool)
	var uniqueItems []interface{}
	for _, item := range this.Items {
		if !uniqueIndexes[item] {
			uniqueIndexes[item] = true
			uniqueItems = append(uniqueItems, item)
		}
	}
	this.Items = uniqueItems
}

// clear is a method of Slice to clear Items
func (this *Slice) Clear() {
	this.Items = []interface{}{}
}

// count is a method of Slice to get count of Items
func (this *Slice) Count() int {
	return len(this.Items)
}

// isEmpty is a method of Slice to check if Items is empty
func (this *Slice) IsEmpty() bool {
	return this.Count() == 0
}

// first is a method of Slice to get first item of Items
func (this *Slice) First() (*interface{}, error) {
	if this.IsEmpty() {
		return nil, errors.New("slice is empty")
	}

	return &this.Items[0], nil
}

// last is a method of Slice to get last item of Items
func (this *Slice) Last() (*interface{}, error) {
	if this.IsEmpty() {
		return nil, errors.New("slice is empty")
	}

	return &this.Items[this.Count()-1], nil
}

// get is a method of Slice to get item at index from Items
func (this *Slice) Get(index int) (interface{}, error) {
	if index < 0 || index >= this.Count() {
		return nil, errors.New("index out of range")
	}

	return this.Items[index], nil
}

func (this *Slice) GetAsPointer(index int) (*interface{}, error) {
	if index < 0 || index >= this.Count() {
		return nil, errors.New("index out of range")
	}

	return &this.Items[index], nil
}

// set is a method of Slice to set item at index to Items
func (this *Slice) Set(index int, item interface{}) {
	// check if index out of range
	if index < 0 || index >= this.Count() {
		return
	}

	this.Items[index] = item
}

// pop is a method of Slice to get and remove last item of Items
func (this *Slice) Pop() (interface{}, error) {
	item, err := this.Last()
	if err != nil {
		return nil, err
	}

	_, err = this.RemoveAt(this.Count() - 1)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// get items as string slice
func (this *Slice) GetItemsAsStringSlice() []string {
	var items []string
	for _, item := range this.Items {
		if s, ok := item.(string); ok {
			items = append(items, s)
		}
	}
	return items
}

// get items as int slice
func (this *Slice) GetItemsAsIntSlice() []int {
	var items []int
	for _, item := range this.Items {
		if i, ok := item.(int); ok {
			items = append(items, i)
		}
	}
	return items
}

// paginate with page and perPage
func (this *Slice) Paginate(page int, perPage int) Slice {
	offset := 0
	if page <= 0 {
		page = 1
	}

	offset = (page - 1) * perPage
	if offset > this.Count() {
		offset = this.Count()
	}

	end := offset + perPage
	if end > this.Count() {
		end = this.Count()
	}

	this.lastStartOffset = offset
	this.lastEndOffset = end
	this.lastPage = page
	return Slice{
		Items: this.Items[offset:end],
	}
}

// get item by new index after Paginate (start from 0, corresponding to offset)
func (this *Slice) GetItemByIndexAfterPaginate(index int, page int, perPage int) interface{} {
	paginated := this.Paginate(page, perPage)
	item, _ := paginated.Get(index)
	return item
}

// get start offset of last paginate
// you need to call Paginate before call this method
func (this *Slice) GetLastStartOffset() int {
	return this.lastStartOffset
}

// get end offset of last paginate
// you need to call Paginate before call this method
func (this *Slice) GetLastEndOffset() int {
	return this.lastEndOffset
}

// get last page of last paginate
// you need to call Paginate before call this method
func (this *Slice) GetLastPage() int {
	return this.lastPage
}

// get total page of given perPage
func (this *Slice) GetTotalPage(perPage int) int {
	totalPage := this.Count() / perPage
	if this.Count()%perPage != 0 {
		totalPage++
	}
	return totalPage
}
