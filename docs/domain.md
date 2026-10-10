# Доменная модель

## Назначение проекта

Проект представляет собой веб-приложение интернет-магазина.

На текущем этапе реализуются две основные бизнес-сущности:

- товар;
- заказ.

Приложение должно предоставлять интерфейс и API для управления этими сущностями.

---

## Текущие ограничения

На данном этапе сознательно не реализуются:

- количество товара в заказе;
- корзина;
- авторизация пользователей;
- пагинация;
- поиск и фильтрация;
- оплата;
- история изменений;
- мягкое удаление сущностей.

Если в будущем потребуется количество товара, модель можно будет адаптировать
через отдельную таблицу `order_items`.

---

## Сущности

### Product

Товар — единица каталога, которую можно добавить в заказ.

Поля:

| Поле  | Тип в API | Тип в Go | Тип в БД | Описание |
|---|---:|---:|---:|---|
| id | integer | int64 | bigserial / bigint | Идентификатор товара |
| name | string | string | text | Уникальное название товара |
| price | integer | int64 | bigint | Цена в минимальных единицах валюты |

#### Правила для Product

1. Название товара не должно быть пустым.
2. Название товара должно быть уникальным.
3. Длина названия товара ограничена.
4. Цена не может быть отрицательной.
5. Товар, который используется хотя бы в одном заказе, нельзя удалить.

Ограничения:

```text
name: required, length 1..255, unique
price: required, integer >= 0
```

Пример:

```json
{
  "id": 1,
  "name": "Ноутбук",
  "price": 10000000
}
```

Если считать, что цена хранится в копейках, то:

```text
10000000 = 100000.00
```

---

### Order

Заказ — заявка покупателя на покупку набора товаров.

Поля:

| Поле  | Тип в API | Тип в Go | Тип в БД | Описание |
|---|---:|---:|---:|---|
| id | integer | int64 | bigserial / bigint | Идентификатор заказа |
| number | string | string | text | Уникальный номер заказа |
| deliveryAddress | string | string | text | Адрес доставки |
| products | array | []Product | связь через order_products | Список товаров в заказе |

#### Правила для Order

1. Номер заказа не должен быть пустым.
2. Номер заказа должен быть уникальным.
3. Адрес доставки не должен быть пустым.
4. Заказ должен содержать хотя бы один товар.
5. Все товары в заказе должны существовать.
6. Товары в заказе не должны повторяться.

Ограничения:

```text
number: required, length 1..64, unique
deliveryAddress: required, length 1..500
productIds: required, min 1, unique elements, existing products
```

Пример:

```json
{
  "id": 1,
  "number": "ORD-0001",
  "deliveryAddress": "Москва, ул. Пушкина, д. 1",
  "products": [
    {
      "id": 1,
      "name": "Ноутбук",
      "price": 10000000
    },
    {
      "id": 2,
      "name": "Мышь",
      "price": 150000
    }
  ]
}
```

---

## Связь заказа и товара

Между `Order` и `Product` используется связь многие ко многим.

Так как количество товара в заказе пока не поддерживается,
связь описывается промежуточной таблицей:

```text
order_products
```

Каждая запись означает:

```text
Товар с указанным product_id входит в заказ с указанным order_id.
```

---

## Маппинг на реляционную базу данных

### products

```text
id      bigint primary key generated always as identity
name    text not null unique
price   bigint not null check (price >= 0)
```

### orders

```text
id               bigint primary key generated always as identity
number           text not null unique
delivery_address text not null
```

### order_products

```text
order_id   bigint not null references orders(id) on delete cascade
product_id bigint not null references products(id) on delete restrict

primary key (order_id, product_id)
```

Пояснение:

```text
Если удаляется заказ, связанные записи из order_products удаляются автоматически.
Если товар используется в заказе, удалить его нельзя.
```

---

## Будущая адаптация под количество товара

Если позже понадобится количество, вместо или вместе с `order_products`
можно ввести таблицу:

```text
order_items
```

Пример:

```text
order_id    bigint not null references orders(id)
product_id  bigint not null references products(id)
quantity    integer not null check (quantity > 0)
```

Текущая модель спроектирована так, чтобы такое расширение можно было добавить
без полного переосмысления домена.