def exchange_money(budget, exchange_rate):
    return budget / exchange_rate


def get_change(budget, exchanging_value):
    return budget - exchanging_value


def get_value_of_bills(denomination, numbers_of_bills):
    return denomination * numbers_of_bills


def get_number_of_bills(amount, denomination):
    return int(amount // denomination)


def get_leftover_of_bills(amount, denomination):
    return amount % denomination


def exchangeable_value(budget, exchange_rate, spread, denomination):
    addition_fee = exchange_rate * (1 + spread / 100)
    money_after_exchange = exchange_money(budget, addition_fee)
    number_of_bills = get_number_of_bills(money_after_exchange, denomination)
    return get_value_of_bills(denomination, number_of_bills)


print(exchangeable_value(127.25, 1.20, 10, 5))
print(exchangeable_value(127.25, 1.20, 10, 20))