from datetime import datetime, timedelta
records = {}


def set(key: str, field: str, value: str) -> None:
    if key in records:
        records[key][field] = value
    else:
        records[key] = {field: value}


def get(key: str, field: str) -> str or None:
    try:
        return records[key][field]
    except KeyError:
        return None


def delete(key: str, field: str) -> bool:
    try:
        del records[key][field]
        return True
    except KeyError:
        return False


def scan(key: str) -> list[str] or []:
    if key not in records:
        return []
    temp = sorted([f'{k}-{v}'for k, v in records[key].items()])
    return temp


def scan_by_prefix(key: str, prefix: str) -> list[str] or []:
    if key not in records:
        return []
    temp = sorted([f'{k}-{value}' for k, value in records[key].items() if k.startswith(prefix)])
    return temp


def set_data_by_timestamp(key: str, field: str, value: str, timestamp: str) -> None:
    dt = datetime.strptime(timestamp, '%Y-%m-%d %H:%M:%S')
    if key in records:
        records[key]['timestamp'] = dt
    else:
        records[key] = {field: value,"timestamp": dt}


def get_data_by_timestamp(key:str, timestamp:str, timelimit: int) -> dict or None:
    dt = datetime.strptime(timestamp, '%Y-%m-%d %H:%M:%S')
    time_in_records = records[key]['timestamp']
    if dt - time_in_records >= timedelta(minutes=timelimit):
        return records[key]
    else:
        return {"error": "Timestamp is less than 5 minutes old"}


def main():
    set("c", "cat", "1")
    set("c", "cow", "2")
    set("a", "ant", "1")
    set("a", "apple", "2")
    # print(get("c", "cat"))
    # print(get("C", "aisle_milk"))
    # print(delete("C", "aisle_milk"))
    # print(delete("C", "tomato"))
    # print(scan("C"))
    # print("prefix : ", scan_by_prefix("A", "app"))
    # print("prefix : ", scan_by_prefix("a", "an"))
    set_data_by_timestamp("a", str(), str(), "2024-06-24 12:00:00")
    print(get_data_by_timestamp("a", "2024-06-24 12:05:00", 5))
    print(records)


if __name__ == "__main__":
    main()