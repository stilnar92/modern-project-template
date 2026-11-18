# Naming Rules by DDD Construct

Read this file when you name a thing of a known `kind:`. It was moved out of `SKILL.md` to keep that file short.

The Index `Kind` column selects the rule.

## Aggregates & Aggregate Roots (`aggregate`)

Use the business domain term. Singular. No technical suffixes.

```text
GOOD: Order, Invoice, UserAccount, ShoppingCart
BAD:  OrderAggregate, OrderRoot, OrderAggregateImpl, OrderEntity
```

### Entities (`entity`)

Singular noun from the domain. Something with identity.

```text
GOOD: OrderLineItem, PaymentTransaction, Customer
BAD:  OrderLineItemEntity, OrderLineItemImpl, OrderLineItemObj
```

### Value Objects (`value`)

Singular noun describing an immutable concept. Describes **what it is**, not what it does.

```text
GOOD: Money, Email, PhoneNumber, Address, DateRange
BAD:  MoneyValue, EmailValidator, PriceInfo, AmountData
```

### Domain Events (`event`)

**Past tense verb + noun.** Something that happened.

```text
GOOD: OrderPlaced, PaymentCaptured, InvoiceSent, InventoryReserved
BAD:  OrderEvent, OnOrderPlaced, CreateOrder (that's a command)
```

### Commands (`command`)

**Imperative verb + noun.** An action requested.

```text
GOOD: CreateOrder, CancelInvoice, ProcessRefund, ReserveInventory
BAD:  OrderCreated (that's an event), NewOrder, OrderCommand
```

### Queries (`query`)

Question or retrieval. Verb + object or descriptive name.

```text
GOOD: GetOrderById, FindInvoicesByCustomer, ListPendingOrders
BAD:  RetrieveOrderData, OrderQuery, GetterForOrder
```

### Domain Services (`service`)

Named after **business activities** the domain expert recognizes.

```text
GOOD: InvoiceCalculator, OrderFulfillment, NotificationSender
BAD:  OrderManager, GenericService, HelperService
```

### Repositories

Repository suffix is acceptable — it's an infrastructure pattern. Repositories are not
thesaurus terms; they take the name of the aggregate they store.

```text
GOOD: OrderRepository, InvoiceRepository, CustomerRepository
BAD:  OrderStorage, OrderPersistence, OrderFinder, OrderDao
```

### Methods on Aggregates

**Commands (change state):** Imperative verb, no "Get" prefix.

```text
GOOD: order.Cancel(), order.AddLineItem(product, quantity), order.Recalculate()
BAD:  order.CancelOrderMethod(), order.GetCancelled(), order.DoCancelOrder()
```

**Queries (read-only):** Start with Get, Is, Has, Can, or a domain verb.

```text
GOOD: order.GetTotal(), order.IsExpired(), order.CanBeShipped()
BAD:  order.FetchInfo(), order.CheckData()
```
