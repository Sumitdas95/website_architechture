def calculate_the_value_of_card(card):
    if card == 'JKQ':
        return 10
    elif card == 'A':
        return 1
    return int(card)


def higher_card(card_one, card_two):
    value_one = calculate_the_value_of_card(card_one)
    value_two = calculate_the_value_of_card(card_two)

    if calculate_the_value_of_card(card_one) == calculate_the_value_of_card(card_two):
        return card_one, card_two
    return max(value_one, value_two)


def value_of_ace(card_one, card_two):
    highest_value = 21
    if card_one == 'A' or card_two == 'A':
        return 1
    total_card_value = calculate_the_value_of_card(card_one) + calculate_the_value_of_card(card_two)

    if highest_value - total_card_value >= 11:
        return 11
    return 1


def is_blackjack(card_one, card_two):
    return calculate_the_value_of_card(card_one) + calculate_the_value_of_card(card_two) == 21


def can_split_pairs(card_one, card_two):
    return calculate_the_value_of_card(card_one) == calculate_the_value_of_card(card_two)


def can_double_down(card_one, card_two):
    total_value = calculate_the_value_of_card(card_one) + calculate_the_value_of_card(card_two)
    return 9 >= total_value <= 11


print(can_double_down('10', '9'))