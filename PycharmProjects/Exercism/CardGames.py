def get_rounds(round_number):
    return [round_number, round_number + 1, round_number + 2]


def concatenate_rounds(round_one, round_two):
    return round_one + round_two


def list_contains_round(rounds, round_number):
    return True if round_number in rounds else False


def card_average(hand):
    return sum(hand) / len(hand)


def approx_average_is_average(hand):
    actual_avg = card_average(hand)
    median = int(len(hand) - 1 / 2)

    return True if actual_avg == hand[median] or actual_avg == (hand[0] + hand[-1]) / 2 else False


def average_even_is_average_odd(hand):
    even_hand = []
    odd_hand = []
    for idx, num in enumerate(hand):
        if idx % 2 == 0:
            even_hand.append(num)
        else:
            odd_hand.append(num)
    return card_average(even_hand) == card_average(odd_hand)


def maybe_double_last(hand):
    if hand[-1] == 11:
        hand[-1] = 22
    return hand