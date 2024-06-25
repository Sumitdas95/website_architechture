Azara = [
            ("Amethyst Octopus", '1F'),
            ("Angry Monkey Figurine", '5B'),
            ("Antique Glass Fishnet Float", '3D'),
            ("Brass Spyglass", '4B'),
            ("Carved Wooden Elephant", '8C'),
            ("Crystal Crab", '6A'),
            ("Glass Starfish", '6D'),
            ("Model Ship in Large Bottle", '8A'),
            ("Pirate Flag", '7F'),
            ("Robot Parrot", '1C'),
            ("Scrimshawed Whale Tooth",	'2A'),
            ("Silver Seahorse",	'4E'),
            ("Vintage Pirate Hat", '7E')
        ]

Rui = [
    ("Seaside Cottages", ("1", "C"), "Blue"),
    ("Aqua Lagoon (Island of Mystery)", ("1", "F"), "Yellow"),
    ("Deserted Docks", ("2", "A"), "Blue"),
    ("Spiky Rocks",	("3", "D"),	"Yellow"),
    ("Abandoned Lighthouse", ("4", "B"), "Blue"),
    ("Hidden Spring (Island of Mystery)", ("4", "E"), "Yellow"),
    ("Stormy Breakwater", ("5", "B"), "Purple"),
    ("Old Schooner", ("6", "A"), "Purple"),
    ("Tangled Seaweed Patch", ("6", "D"), "Orange"),
    ("Quiet Inlet (Island of Mystery)",	("7", "E"),	"Orange"),
    ("Windswept Hilltop (Island of Mystery)", ("7", "F"), "Orange"),
    ("Harbor Managers Office", ("8", "A"),	"Purple"),
    ("Foggy Seacave", ("8", "C"), "Purple")
]


def get_coordinate(treasure, coordinate):
    if (treasure, coordinate) in Azara:
        return coordinate


def convert_coordinate(coordinate):
    return coordinate[0], coordinate[1]


def compare_records(loc_coordinate_one, loc_coordinate_two):
    coordinate_refactor = convert_coordinate(loc_coordinate_one[-1])
    return coordinate_refactor == loc_coordinate_two[-2]


def create_record(loc_coordinate_one, loc_coordinate_two):
    coordinate_refactor = convert_coordinate(loc_coordinate_one[-1])
    return loc_coordinate_one + loc_coordinate_two if coordinate_refactor == loc_coordinate_two[-2] else "not a match"


def clean_up(records):
    clean_up_records = ()
    for record in records:
        list_of_records = []
        list_of_records.append(record[0])
        list_of_records += record[2:]
        clean_up_records += tuple(list_of_records)
    return clean_up_records


# print(compare_records(('Model Ship in Large Bottle', '8A'), ('Harbor Managers Office', ('8', 'A'), 'purple')))
# print(compare_records(('Brass Spyglass', '4B'), ('Seaside Cottages', ('1', 'C'), 'blue')))
# print(create_record(('Brass Spyglass', '4B'), ('Abandoned Lighthouse', ('4', 'B'), 'Blue')))
# print(create_record(('Brass Spyglass', '4B'), ('Seaside Cottages', ('1', 'C'), 'blue')))
print(clean_up((('Brass Spyglass', '4B', 'Abandoned Lighthouse', ('4', 'B'), 'Blue'), ('Vintage Pirate Hat', '7E', 'Quiet Inlet (Island of Mystery)', ('7', 'E'), 'Orange'), ('Crystal Crab', '6A', 'Old Schooner', ('6', 'A'), 'Purple'))))