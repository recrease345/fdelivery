| Судьба | Что |
|---|---|
| **Переезжают как есть** | TestNewOrderSuccess, TestNewOrderValidation (13 кейсов), TestNewOrderTotalPrice, TestCompensationFor-умрёт (мёртвый код) |
| **Мелкая правка** | `newOrderInStatus` → `Restore` (3 строки) |
| **Переписать** | Матрица: с `transitionTo` на публичные методы — и это единственная настоящая работа |
| **Новые** | контракт К3, фикс К1, инвариант компенсаций (у тебя уже есть) |

package order

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID          = errors.New("order: userID is nil")
	ErrInvalidItems           = errors.New("order: items must be non-empty")
	ErrInvalidRestaurantID    = errors.New("order: restaurantID is nil")
	ErrInvalidItemsAttributes = errors.New("order: invalid item attributes")
	ErrInvalidDeliveryAddress = errors.New("order: invalid delivery address")
	ErrInvalidStatus          = errors.New("order: restore with invalid status")
	ErrInvalidID              = errors.New("order: id is nil")
)

type Item struct {
	DishID    uuid.UUID
	Name      string
	UnitPrice int64
	Quantity  int
}

type Order struct {
	id              uuid.UUID
	userID          uuid.UUID
	restaurantID    uuid.UUID
	items           []Item
	totalPrice      int64
	status          Status
	deliveryAddress string
	createdAt       time.Time
}

func NewOrder(userID uuid.UUID, restaurantID uuid.UUID, items []Item, deliveryAddress string, now time.Time) (*Order, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidID
	}

	if restaurantID == uuid.Nil {
		return nil, ErrInvalidRestaurantID
	}

	if len(items) == 0 {
		return nil, ErrInvalidItems
	}

	if deliveryAddress == "" {
		return nil, ErrInvalidDeliveryAddress
	}

	var totalPrice int64
	for _, item := range items {
		if item.DishID == uuid.Nil || item.Name == "" || item.Quantity <= 0 || item.UnitPrice <= 0 {
			return nil, ErrInvalidItemsAttributes
		}

		totalPrice += (item.UnitPrice * int64(item.Quantity))
	}

	return &Order{
		id:              uuid.New(),
		userID:          userID,
		restaurantID:    restaurantID,
		items:           slices.Clone(items),
		totalPrice:      totalPrice,
		status:          Created,
		deliveryAddress: deliveryAddress,
		createdAt:       now,
	}, nil
}

func Restore(id uuid.UUID, userID uuid.UUID, restaurantID uuid.UUID, totalPrice int64, items []Item, deliveryAddress string, createdAt time.Time, status Status) (*Order, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	switch status {
	case Created, Paid, Preparing, Ready, Assigning, Assigned, Delivering, Delivered, Cancelled, FailedAssign:
	default:
		return nil, ErrInvalidStatus
	}

	return &Order{
		id:              id,
		userID:          userID,
		restaurantID:    restaurantID,
		items:           slices.Clone(items),
		totalPrice:      totalPrice,
		status:          status,
		deliveryAddress: deliveryAddress,
		createdAt:       createdAt,
	}, nil
}

func (o *Order) cancel() (Compensation, *InvalidTransitionError) {
	сomp := compensations[o.status]
	if err := o.transitionTo(Cancelled); err != nil {
		return CompUnset, err
	}

	return сomp, nil
}

func (o *Order) paid() *InvalidTransitionError {
	if err := o.transitionTo(Paid); err != nil {
		return err
	}

	return nil
}

func (o *Order) preparing() *InvalidTransitionError {
	if err := o.transitionTo(Preparing); err != nil {
		return err
	}

	return nil
}

func (o *Order) ready() *InvalidTransitionError {
	if err := o.transitionTo(Ready); err != nil {
		return err
	}

	return nil
}

func (o *Order) assigning() *InvalidTransitionError {
	if err := o.transitionTo(Assigning); err != nil {
		return err
	}

	return nil
}

func (o *Order) assigned() *InvalidTransitionError {
	if err := o.transitionTo(Assigned); err != nil {
		return err
	}

	return nil
}

func (o *Order) delivering() *InvalidTransitionError {
	if err := o.transitionTo(Delivering); err != nil {
		return err
	}

	return nil
}

func (o *Order) delivered() *InvalidTransitionError {
	if err := o.transitionTo(Delivered); err != nil {
		return err
	}

	return nil
}

func (o *Order) failedAssign() (Compensation, *InvalidTransitionError) {
	Comp := compensations[o.status]

	if err := o.transitionTo(FailedAssign); err != nil {
		return CompUnset, err
	}

	return Comp, nil
}

// getters

func (o *Order) ID() uuid.UUID {
	return o.id
}

func (o *Order) UserID() uuid.UUID {
	return o.userID
}

func (o *Order) RestaurantID() uuid.UUID {
	return o.restaurantID
}

func (o *Order) TotalPrice() int64 {
	return o.totalPrice
}

func (o *Order) DeliveryAddress() string {
	return o.deliveryAddress
}

func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Order) Status() Status {
	return o.status
}

func (o *Order) Items() []Item {
	return slices.Clone(o.items)
}
package order

import (
	"fmt"
)

type Status string

const (
	Created      Status = "created"
	Paid         Status = "paid"
	Preparing    Status = "preparing"
	Ready        Status = "ready"
	Assigning    Status = "assigning"
	Assigned     Status = "assigned"
	Delivering   Status = "delivering"
	Delivered    Status = "delivered"
	Cancelled    Status = "cancelled"
	FailedAssign Status = "failed_assign"
)

type InvalidTransitionError struct {
	From, To Status
}

func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("order: invalid transition %s -> %s", e.From, e.To)
}

var transitions = map[Status][]Status{
	Created:    {Paid, Cancelled},
	Paid:       {Preparing, Cancelled},
	Preparing:  {Ready, Cancelled},
	Ready:      {Assigning, Cancelled},
	Assigning:  {Assigned, FailedAssign, Cancelled},
	Assigned:   {Delivering, Cancelled},
	Delivering: {Delivered},
}

type Compensation int

const (
	CompUnset Compensation = iota
	CompNone  
	CompVoid
	CompVoidRestaurant
	CompVoidReleaseDriverRestaurant
)

var compensations = map[Status]Compensation{
	Created:   CompNone,
	Paid:      CompVoid,
	Preparing: CompVoidRestaurant,
	Ready:     CompVoidRestaurant,
	Assigning: CompVoidReleaseDriverRestaurant,
	Assigned:  CompVoidReleaseDriverRestaurant,
}

func (order *Order) transitionTo(to Status) *InvalidTransitionError {
	for _, status := range transitions[order.status] {
		if status == to {
			order.status = to
			return nil
		}
	}

	return &InvalidTransitionError{
		From: order.status,
		To:   to,
	}
}
это домен, теперь тесты старые: package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompensationFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status Status
		want   Compensation
	}{
		{"created", Created, CompNone},
		{"paid", Paid, CompVoid},
		{"preparing", Preparing, CompVoidRestaurant},
		{"ready", Ready, CompVoidRestaurant},
		{"assigning", Assigning, CompVoidReleaseDriverRestaurant},
		{"assigned", Assigned, CompVoidReleaseDriverRestaurant},
		{"delivering", Delivering, CompForbidden},
		{"terminal cancelled", Cancelled, CompUnset},
		{"terminal delivered", Delivered, CompUnset},
		{"terminal failed assign", FailedAssign, CompUnset},
		{"unknown status", Status("unknownStatus"), CompUnset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, CompensationFor(tt.status))
		})
	}
}

func TestEveryCancellableStatusHasCompensation(t *testing.T) {
	t.Parallel()

	for from, targets := range transitions {
		for _, to := range targets {
			if to == Cancelled && Compensations[from] == CompUnset {
				t.Errorf("status %q is cancellable but Compensation policy is not defined", from)
			}
		}
	}
}
package order

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

func validItem() Item {
	return Item{
		DishID:    uuid.New(),
		Name:      "пицца",
		UnitPrice: 59900,
		Quantity:  3,
	}
}

func anotherItem() Item {
	return Item{
		DishID:    uuid.New(),
		Name:      "тройной чизбургер дабл трипл воппер",
		UnitPrice: 29900,
		Quantity:  1,
	}
}

// изначально валидна, потом по одному полю ломается
type fixture struct {
	userID          uuid.UUID
	restaurantID    uuid.UUID
	items           []Item
	deliveryAddress string
	now             time.Time
}

func validFixture() fixture {
	items := []Item{validItem(), anotherItem()}

	return fixture{
		userID:          uuid.New(),
		restaurantID:    uuid.New(),
		items:           items,
		deliveryAddress: "Улица Пушкина Дом колотушкина",
		now:             fixedNow,
	}
}

func newOrder(t testing.TB) *Order {
	t.Helper()

	f := validFixture()
	o, err := NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, f.now)
	require.NoError(t, err, "fixture: valid data must not cause an error")
	return o
}

func newOrderInStatus(t testing.TB, s Status) *Order {
	t.Helper()

	o := newOrder(t)
	o.Status = s
	return o
}

func allStatuses() []Status {
	return []Status{
		Created, Paid, Preparing, Ready, Assigning, Assigned, Delivering, Delivered, Cancelled, FailedAssign,
	}
}
package order

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrderSuccess(t *testing.T) {
	t.Parallel()

	f := validFixture()

	got, err := NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, fixedNow)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, f.userID, got.UserID)
	assert.Equal(t, f.restaurantID, got.RestaurantID)
	assert.Equal(t, f.items, got.Items)
	assert.Equal(t, f.deliveryAddress, got.DeliveryAddress)
	assert.Equal(t, f.now, got.CreatedAt)
	assert.Equal(t, Created, got.Status)
	assert.Equal(t, int64(209600), got.TotalPrice, "59900*3 + 29900*1")
	assert.NotEqual(t, uuid.Nil, got.ID)
}

func TestNewOrderValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(f *fixture)
		wantErr error
	}{
		{"nil user id", func(f *fixture) { f.userID = uuid.Nil }, ErrInvalidUserID},
		{"nil restaurant id", func(f *fixture) { f.restaurantID = uuid.Nil }, ErrInvalidRestaurantID},
		{"nil items", func(f *fixture) { f.items = nil }, ErrInvalidItems},
		{"empty items", func(f *fixture) { f.items = []Item{} }, ErrInvalidItems},
		{"empty delivery address", func(f *fixture) { f.deliveryAddress = "" }, ErrInvalidDeliveryAddress},
		{"item empty name", func(f *fixture) { f.items[0].Name = "" }, ErrInvalidItemsAttributes},
		{"item nil dish id", func(f *fixture) { f.items[0].DishID = uuid.Nil }, ErrInvalidItemsAttributes},
		{"item zero quantity", func(f *fixture) { f.items[0].Quantity = 0 }, ErrInvalidItemsAttributes},
		{"item negative quantity", func(f *fixture) { f.items[0].Quantity = -1 }, ErrInvalidItemsAttributes},
		{"item zero unit price", func(f *fixture) { f.items[0].UnitPrice = 0 }, ErrInvalidItemsAttributes},
		{"item negative unit price", func(f *fixture) { f.items[0].UnitPrice = -1 }, ErrInvalidItemsAttributes},
		{"second item invalid", func(f *fixture) { f.items[1].Quantity = 0 }, ErrInvalidItemsAttributes},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := validFixture()
			tt.mutate(&f)

			got, err := NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, fixedNow)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, got, "on validation error constructor must return nil")
		})
	}
}

func TestNewOrderItemsAreDefensiveCopy(t *testing.T) {
	t.Parallel()

	f := validFixture()

	o, err := NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, fixedNow)
	require.NoError(t, err)

	want := f.items[0].Quantity
	f.items[0].Quantity = -100

	assert.Equal(t, want, o.Items[0].Quantity, "mutation of the callers slice should not change it")
}

func TestNewOrderTotalPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		items []Item
		want  int64
	}{
		{
			"qty multiplication",
			[]Item{{DishID: uuid.New(), Name: "крылышки кфс", UnitPrice: 10000, Quantity: 2}},
			20000,
		},
		{
			"items accumulate",
			[]Item{validItem(), anotherItem()},
			209600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o, err := NewOrder(uuid.New(), uuid.New(), tt.items, "Улица пушкина дом колотушкина", fixedNow)

			require.NoError(t, err)
			assert.Equal(t, tt.want, o.TotalPrice)
		})
	}
}
package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allowedTransitions = map[Status]map[Status]bool{
	Created:    {Paid: true, Cancelled: true},
	Paid:       {Preparing: true, Cancelled: true},
	Preparing:  {Ready: true, Cancelled: true},
	Ready:      {Assigning: true, Cancelled: true},
	Assigning:  {Assigned: true, FailedAssign: true, Cancelled: true},
	Assigned:   {Delivering: true, Cancelled: true},
	Delivering: {Delivered: true},
}

func TestTransitionTo_Matrix(t *testing.T) {
	t.Parallel()

	for _, from := range allStatuses() {
		for _, to := range allStatuses() {
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				t.Parallel()

				o := newOrderInStatus(t, from)

				err := o.transitionTo(to)

				if allowedTransitions[from][to] {
					require.Nil(t, err)
					assert.Equal(t, to, o.Status, "successful transition should change the status")
				} else {
					require.NotNil(t, err)
					assert.Equal(t, &InvalidTransitionError{From: from, To: to}, err,
						"error must contain both statuses")
					assert.Equal(t, from, o.Status, "an unsuccessful transition should not change the status")
				}
			})
		}
	}
}

func TestTransitionToFullLifeCycle(t *testing.T) {
	t.Parallel()

	o := newOrder(t)

	for _, s := range []Status{
		Paid, Preparing, Ready, Assigning, Assigned, Delivering, Delivered,
	} {
		require.Nil(t, o.transitionTo(s), "transition to %s should be allowed", s)
	}

	assert.Equal(t, Delivered, o.Status)
}

func TestTransitionToUnknownTargetStatus(t *testing.T) {
	t.Parallel()

	o := newOrder(t)

	err := o.transitionTo(Status("unknownStatus"))

	require.NotNil(t, err)

	assert.Equal(t, &InvalidTransitionError{From: Created, To: Status("unknownStatus")}, err)
}
там много ошибок тк они для старого домена, со всеми обьяснениями(особенно testify) адаптируй их под новый домен по таблице с начала