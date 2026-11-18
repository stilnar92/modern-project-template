# Naming Anti-Patterns to Detect and Flag

Read this file when you check a name or review a naming change. It was moved out of `SKILL.md` to keep that file short.

## Lexical Firewall: the `## Forbidden` section

The domain layer must be protected from transient jargon, vague terms, and
implementation details. The thesaurus's `## Forbidden` section lists words that MUST NOT
appear in domain names and must always be replaced with a specific domain term.

### Weasel Words (never use in domain layer)

| Weasel Word | Problem | Fix |
|-------------|---------|-----|
| `Info` | Meaningless suffix | Remove it: `UserInfo` -> `User` |
| `Data` | Says nothing about the concept | Use domain term: `OrderData` -> `Order` |
| `Manager` | Vague, hides responsibility | Split by actual responsibility |
| `Handler` | Generic, unclear intent | Name after what it handles |
| `Service` | Overused catch-all | Use specific domain activity name |
| `Base` | Technical distraction | Remove, use composition |
| `Item` | Too generic | Use domain term: `Item` -> `OrderLineItem`, `Product` |
| `Util` / `Helper` | Indicates bad design | Move logic to domain objects |
| `Object` / `Obj` | Never appropriate | Remove suffix |
| `Record` / `Model` | Database concept leaking into domain | Use domain term |
| `Config` / `Settings` | Generic container hiding a concept | `Config` -> `LoanProduct`, `Settings` -> `NotificationPreferences` |

### Technical Jargon in Domain Layer

Domain code must be free of implementation details:

```text
BAD:  MongoOrder, SqlUserRepository, HttpOrderService, OrderDto, OrderEntity
GOOD: Order, OrderRepository (interface), PaymentGateway, Order (just Order)
```

Technical prefixes/suffixes belong ONLY in the infrastructure layer, and even there the
domain role should lead:

```text
INFRASTRUCTURE LAYER (OK): MongoOrderRepository, RedisSessionCache, HttpPaymentClient
DOMAIN LAYER (NEVER):      MongoOrder, RedisSession, HttpPayment
NAME THE ROLE, NOT THE TECH: SessionStore not RedisCache, EventPublisher not KafkaProducer
```

**Framework caveat**: In frameworks that intentionally blend domain and persistence
(Active Record pattern, ORM-centric frameworks), the model IS the domain entity.
Keep the **domain noun clean** and let framework coupling live in inheritance,
annotations, or metadata — not in the class name. Flag technical jargon only when
it becomes part of the business-facing name or leaks outside its boundary.

### Synonym Drift

Same concept called different things in different parts of code:

```text
PROBLEM: "Customer" in auth, "User" in API, "Account" in billing — all mean the same thing
FIX:     Pick ONE canonical term per bounded context. Put the others in that line's `avoid:` list.
```

### Abbreviation Boundary

Ban abbreviations in **durable, domain-bearing names**: types, exported functions,
modules, API fields, DB columns, events, config keys.

Allow **conventional short-lived local identifiers** when meaning is obvious in scope:
`i`, `j`, `ctx`, `req`, `res`, `err`, `tx`, `db`, `e` for events.

Allow **industry-standard acronyms** when they are the dominant term: `SKU`, `VAT`,
`URL`, `ID`, `OAuth`. Do NOT force unnatural expansions if experts use the acronym.

```text
PROBLEM: usr, user, account, acct — competing abbreviations for the same durable concept
FIX:     Pick ONE canonical form for domain-bearing names. Short-lived locals are exempt.
```

### Translation Chain ("Telephone Game")

When different artifacts use different terms for the same concept across the
knowledge chain, information is lost at each translation:

```text
SMELL: Domain expert says "Campaign" → PM writes "Promotion" in spec →
       Dev codes `marketing_push` → QA tests "advertising effort"
FIX:   Same term everywhere: expert, PM, dev, QA all say and write "Campaign"
```

This is worse than synonym drift because each translation also loses nuance and
business rules. **How to detect:** compare terms in requirements/specs/tickets
against code names. If they don't match, the ubiquitous language has a translation
gap — adopt the domain expert's term everywhere.
