// Primers match a deck by name. No entry means the How to play page stays empty.
// level.scale is "Bracket" for Commander and "Kitchen" for 60-card.
// Kitchen is not a Standard tier: Standard ranks tournament decks Tier 1–3.
const deckGuides = {
  "Кумена — Секретные потоки": {
    kicker: "60 карт · синий и зелёный · 20 жизней",
    hero: "Kumena, Tyrant of Orazca",
    level: {
      scale: "Kitchen",
      name: "Focused",
      note: "Один план, и он приходит почти каждую игру. Не тир Standard: тирами меряют турнирную мету."
    },
    lead: "Мерфолки выходят рано, на них садятся counters, и Herald of Secret Streams проводит их мимо блоков. Kumena с этого стола берёт карты.",
    stats: [
      { k: "Elite", v: "ход 2" },
      { k: "Counter", v: "×2" },
      { k: "Draw", v: "tap 3" },
      { k: "Мимо блоков", v: "unblockable" }
    ],
    turnsLine: "К пятому ходу counters уже на столе. Herald проводит этих существ, а Kumena тратит tap трёх мерфолков и берёт карту.",
    turns: [
      { n: "1", card: "Llanowar Elves", tag: "{G}", text: "Зелёная мана. На втором ходу хватает на Elite." },
      { n: "2", card: "Deeproot Elite", tag: "+1/+1", text: "Ждёт. Следующий мерфолк кладёт кому-то counter." },
      { n: "3", card: "Jungleborn Pioneer", tag: "token", text: "Token 1/1 с hexproof. Elite кладёт counter на token или на себя." },
      { n: "4", card: "Herald of Secret Streams", tag: "unblockable", text: "У кого есть +1/+1 counter, тот проходит." },
      { n: "5", card: "Kumena, Tyrant of Orazca", tag: "draw", text: "Tap трёх мерфолков. Draw." }
    ],
    sections: [
      {
        title: "Кто выходит первым",
        paragraphs: ["На одной мане есть выбор. Эльф ускоряет. Speaker становится 2/2, если рядом мерфолк или остров. Herald и так не блокируется."],
        beats: [
          { card: "Llanowar Elves", stat: "{G}", text: "Tap: добавь {G}. Без него Elite часто опаздывает на ход." },
          { card: "Kumena's Speaker", stat: "2/2", text: "+1/+1, пока контролируешь другого мерфолка или остров." },
          { card: "Mist-Cloaked Herald", stat: "1 мана", text: "Это существо нельзя заблокировать. Одна мана и атака каждый ход, даже без counters." }
        ]
      },
      {
        title: "Откуда счётчики",
        paragraphs: ["Mistbinder даёт остальным мерфолкам статичные +1/+1. Это не counters, Herald of Secret Streams их не видит. Линия ниже как раз про counters."],
        plays: [{
          line: "Elite кладёт один +1/+1 counter. Michelangelo делает из него два. Herald проводит этих существ мимо блоков.",
          ticks: ["мерфолк выходит", "Elite: +1/+1", "Michelangelo: ещё один", "Herald: unblockable"],
          steps: [
            { card: "Deeproot Elite", via: "+1/+1", text: "Чужой мерфолк вышел. Один counter на любого мерфолка. В колоде три копии, одна foil." },
            { card: "Michelangelo, Weirdness to 11", via: "×2", text: "Если на существо кладут counters, их на один больше. Counter от Elite становится двумя." },
            { card: "Herald of Secret Streams", via: "unblockable", text: "Существа с +1/+1 counter нельзя заблокировать. Две копии." }
          ]
        }],
        beats: [
          { card: "Merfolk Mistbinder", stat: "не counter", text: "Остальные мерфолки просто +1/+1. Блок они всё равно получают, Herald их не проводит." },
          { card: "Silvergill Adept", stat: "draw", text: "При розыгрыше покажи мерфолка из руки или доплати {3}. При выходе draw." }
        ]
      },
      {
        title: "Когда стол уже есть",
        paragraphs: ["Kumena сама почти не бьёт. Она тратит untap мерфолков: один делает её unblockable, три дают карту, пять раздают counters."],
        plays: [{
          line: "Maskwood Nexus записывает в мерфолки вообще всех. После этого эльф, tokens и Kumena смотрят на один стол, и tap 5 уже реально собрать.",
          steps: [
            { card: "Maskwood Nexus", via: "все merfolk", text: "Все существа становятся мерфолками. Elite и Kopala тоже начинают их видеть." },
            { card: "Kumena, Tyrant of Orazca", via: "tap 3 / 5", text: "Три untap: карта. Пять untap: +1/+1 каждому мерфолку, Michelangelo удваивает." },
            { card: "Herald of Secret Streams", via: "unblockable", text: "Counters с пятого режима Kumena сразу проводят весь стол." }
          ]
        }],
        beats: [
          { card: "Kopala, Warden of Waves", stat: "+{2}", text: "Чужое заклинание или способность по твоему мерфолку стоит на {2} дороже." },
          { card: "Jungleborn Pioneer", stat: "1/1 hexproof", text: "При выходе token мерфолк. Elite может положить counter на него, и снять token уже непросто." }
        ]
      },
      {
        title: "Что держать в руке",
        paragraphs: ["Контры бьют по нематериальным заклинаниям. Снятие существа здесь одно, и оно ещё и draw."],
        beats: [
          { card: "Spell Pierce", stat: "×2", text: "Несущество контрится, если соперник не доплатит {2}." },
          { card: "Unwind", stat: "×2", text: "Контр несущества, затем untap до трёх земель и продолжай ход." },
          { card: "Crashing Tide", stat: "flash", text: "Верни существо и draw. Пока есть мерфолк, у Tide есть flash." },
          { card: "Blossoming Defense", stat: "+2/+2", text: "Своему существу +2/+2 и hexproof до конца хода. Ответ на точечное снятие." },
          { card: "Heritage Reclamation", stat: "graveyard", text: "Артефакт или чары. Либо карта из graveyard, и тогда ещё draw." },
          { card: "Origin of Metalbending", stat: "shield", text: "Сними артефакт или чары. Либо дай своему существу counter и indestructible до конца хода." }
        ]
      }
    ]
  },
  "Cloud — Final Fantasy Starter": {
    kicker: "60 карт · красный и белый · 20 жизней",
    hero: "Cloud, Planet's Champion",
    level: {
      scale: "Kitchen",
      name: "Starter",
      note: "Стартовый набор: один план, без запаса копий. Не тир Standard."
    },
    lead: "Оружие само создаёт Hero 1/1 и надевается на него. Пока Cloud экипирован, в твой ход у него double strike и indestructible, а надевать на него дешевле.",
    stats: [
      { k: "Hero", v: "1/1" },
      { k: "Equip Cloud", v: "−{2}" },
      { k: "В твой ход", v: "×2" },
      { k: "Пока одет", v: "indestructible" }
    ],
    turnsLine: "Копьё и катана сами сажают героев. На четвёртом ходу оружие переезжает на Cloud, и он бьёт дважды.",
    turns: [
      { n: "2", card: "Dragoon's Lance", tag: "Hero", text: "Hero 1/1 выходит уже с копьём. В твой ход у него flying." },
      { n: "3", card: "Samurai's Katana", tag: "haste", text: "Второй герой: +2/+2, trample и haste." },
      { n: "4", card: "Cloud, Planet's Champion", tag: "−{2}", text: "Надень катану на Cloud за {3} вместо {5}. В твой ход double strike и indestructible." },
      { n: "5", card: "Judgment Bolt", tag: "5+", text: "5 damage существу и ещё игроку за каждую экипировку на board." }
    ],
    sections: [
      {
        title: "Оружие",
        paragraphs: ["Job select один и тот же: при выходе token Hero 1/1, и экипировка уже на нём. Обычный equip дорогой, поэтому второе оружие лучше вешать на Cloud."],
        beats: [
          { card: "Dragoon's Lance", stat: "+1/+0", text: "Тип Knight и flying в твой ход. Equip {4}, на Cloud выходит {2}." },
          { card: "Samurai's Katana", stat: "+2/+2", text: "Trample, haste и тип Samurai. Этот герой бьёт в тот же ход, как вышел." },
          { card: "Warrior's Sword", stat: "+3/+2", text: "Самый толстый из дешёвых клинков. Equip {5}, на Cloud {3}." },
          { card: "White Mage's Staff", stat: "1 life", text: "+1/+1, 1 жизнь за атаку, тип Cleric." }
        ]
      },
      {
        title: "Кто носит оружие",
        paragraphs: ["Cloud выигрывает именно тот ход, когда на нём лежит экипировка. Beatrix перекладывает уже надетое оружие, не платя equip."],
        plays: [{
          line: "Копьё само создаёт Hero и вешается на него. В начале боя Beatrix снимает оружие с героев и надевает на Cloud. Equip за это не платят, а Cloud в твой ход бьёт дважды и его не снять.",
          ticks: ["оружие: Hero 1/1", "Beatrix перевешивает", "Cloud: −{2} на новый equip", "double strike"],
          steps: [
            { card: "Dragoon's Lance", via: "Hero 1/1", text: "Token выходит уже экипированным. Отдельно платить {4} не нужно." },
            { card: "Beatrix, Loyal General", via: "перевесь", text: "В начале боя перекладывает любое число экипировок на одно существо." },
            { card: "Cloud, Planet's Champion", via: "double strike", text: "В твой ход с оружием: double strike и indestructible. Новый equip дешевле на {2}." }
          ]
        }],
        beats: [
          { card: "Adelbert Steiner", stat: "lifelink", text: "+1/+1 за каждую экипировку на столе, не за каждое существо. Lifelink." },
          { card: "Freya Crescent", stat: "{R}", text: "В твой ход flying. Tap даёт {R} только на заклинание экипировки или на equip." }
        ]
      },
      {
        title: "Снятие и конец игры",
        paragraphs: ["Урон и снятие здесь считаются от оружия на столе, а не от голой силы существа."],
        plays: [{
          line: "Каждая экипировка добавляет урон в Bolt и в Slash. Ultima на атаке просто снимает чужое существо и даёт носителю +7/+7.",
          steps: [
            { card: "Barret Wallace", via: "за каждое", text: "Сколько существ одето, столько урона игрок и получает на атаке." },
            { card: "Judgment Bolt", via: "5 + equip", text: "5 damage существу. Игроку ещё по одному за каждую экипировку." },
            { card: "Ultima Weapon", via: "+7/+7", text: "На атаке destroy чужое существо. Носитель получает +7/+7." }
          ]
        }],
        beats: [
          { card: "Slash of Light", stat: "существа + оружие", text: "Урон существу равен числу твоих существ плюс число экипировок." },
          { card: "White Auracite", stat: "exile", text: "Изгоняет чужую неземлю, пока сама на столе, и даёт {W}." },
          { card: "Lightning, Security Sergeant", stat: "exile cast", text: "Урон игроку изгоняет верх библиотеки. Эту карту можно разыграть, пока Lightning на столе." }
        ]
      }
    ]
  },
  "Sephiroth — Final Fantasy Starter": {
    kicker: "60 карт · синий и чёрный · 20 жизней",
    hero: "Sephiroth, Planet's Heir",
    level: {
      scale: "Kitchen",
      name: "Starter",
      note: "Стартовый набор: снятие и graveyard. Не тир Standard."
    },
    lead: "Sephiroth при выходе даёт чужим существам −2/−2. Мелочь умирает, и каждая такая смерть делает его больше. Остальная колода добивает крупных и поднимает своих из graveyard.",
    stats: [
      { k: "При выходе", v: "−2/−2" },
      { k: "Чужая смерть", v: "+1/+1" },
      { k: "Intervention", v: "×3" },
      { k: "Из graveyard", v: "+2/+2" }
    ],
    turnsLine: "Graveyard наполняется раньше командира. На шестом ходу −2/−2 снимает мелочь, Sephiroth растёт, и следующим ходом свой выходит из graveyard ещё крупнее.",
    turns: [
      { n: "3", card: "Shinra Reinforcements", tag: "mill 3", text: "Mill 3 и gain 3 life. Graveyard начинается раньше Sephiroth." },
      { n: "4", card: "Sephiroth's Intervention", tag: "destroy", text: "Destroy существо, gain 2 life. В колоде три копии." },
      { n: "6", card: "Sephiroth, Planet's Heir", tag: "−2/−2", text: "Чужие существа −2/−2. Кто умер, тот кладёт на Sephiroth +1/+1." },
      { n: "7", card: "Evil Reawakened", tag: "+2/+2", text: "Существо из graveyard выходит с двумя лишними +1/+1." }
    ],
    sections: [
      {
        title: "Sephiroth и снятие",
        paragraphs: ["Сначала умирает мелочь от самого прихода Sephiroth. Крупных добирают Intervention и Overkill. Каждая чужая смерть кормит командира."],
        plays: [{
          line: "Intervention снимает существо ещё до командира. Sephiroth добивает оставшихся −2/−2 и кладёт себе +1/+1 за каждую смерть. Evil Reawakened возвращает своего уже с двумя counters.",
          ticks: ["Intervention: destroy", "приход: −2/−2", "смерть: +1/+1", "из graveyard: +2/+2"],
          steps: [
            { card: "Sephiroth's Intervention", via: "destroy", text: "Три копии. Существо умирает, ты получаешь 2 жизни." },
            { card: "Sephiroth, Planet's Heir", via: "−2/−2", text: "Vigilance. Мелочь не переживает приход. Дальше он растёт с каждой чужой смерти." },
            { card: "Evil Reawakened", via: "+2/+2", text: "Своё существо из graveyard выходит с двумя лишними +1/+1." }
          ]
        }],
        beats: [
          { card: "Overkill", stat: "×2", text: "Просто destroy существо за {2}{B}. Когда Intervention жалко тратить." },
          { card: "Deadly Embrace", stat: "draw", text: "Destroy чужое существо, потом draw за каждое существо, умершее в этот ход." }
        ]
      },
      {
        title: "Graveyard",
        paragraphs: ["Отсюда карты не комбо сами с собой. Они либо наполняют стопку, либо возвращают то, что уже умерло."],
        beats: [
          { card: "Dreams of Laguna", stat: "из yard ещё раз", text: "Посмотри верх, можешь discard в graveyard, потом draw. Из graveyard кастуется ещё раз за {3}{U} и потом изгоняется." },
          { card: "Evil Reawakened", stat: "+2/+2", text: "Существо из graveyard выходит с двумя лишними counters. Это основной подъём колоды." },
          { card: "Xande, Dark Mage", stat: "+1/+1 за карту", text: "Menace и +1/+1 за каждую несущество и неземлю в graveyard." },
          { card: "Shambling Cie'th", stat: "из yard в руку", text: "Выходит tapped. Возвращается в руку из graveyard, когда кастуешь несущество и платишь {B}." }
        ]
      },
      {
        title: "Чем заканчивать",
        beats: [
          { card: "Ultimecia, Temporal Threat", stat: "tap всех", text: "При выходе поворачивает всех чужих. Урон твоего существа по игроку draw." },
          { card: "Relm's Sketching", stat: "копия", text: "Копия чужого или своего артефакта, существа или земли." },
          { card: "Al Bhed Salvagers", stat: "−1 / +1", text: "Смерть твоего существа или артефакта: соперник теряет 1, ты получаешь 1." },
          { card: "Magitek Scythe", stat: "+2/+1", text: "Можно сразу надеть: в этот ход first strike, и существо обязаны блокировать." }
        ]
      }
    ]
  },
  "Terra — Revival Trance": {
    kicker: "Commander · красный, белый, чёрный · 100 карт · 40 жизней",
    hero: "Terra, Herald of Hope",
    level: {
      scale: "Bracket",
      rank: "2",
      name: "Core",
      note: "Прекон. Победа собирается на столе и её видно заранее. Двухкарточного комбо и быстрых Game Changers нет, партия живёт около восьми ходов."
    },
    lead: "Terra в начале боя mill две карты и до конца хода flying. Если она ударила игрока, за {2} можно вернуть tapped существо с power 3 или меньше. Крупных поднимают уже отдельные заклинания.",
    stats: [
      { k: "Каждый бой", v: "mill 2" },
      { k: "Return", v: "{2}" },
      { k: "Power", v: "≤ 3" },
      { k: "Входит", v: "tapped" }
    ],
    turnsLine: "К пятому ходу graveyard уже рабочий: Supplier и Terra накидали карт, Anger даёт haste всему столу, Reanimate достаёт крупное за жизни.",
    turns: [
      { n: "2", card: "Stitcher's Supplier", tag: "mill 3", text: "Mill 3 сразу и ещё mill 3, когда он умрёт." },
      { n: "3", card: "Terra, Herald of Hope", tag: "mill 2", text: "Commander на board. Со следующего боя mill идёт сам." },
      { n: "4", card: "Anger", tag: "haste", text: "Лучше в graveyard, чем на board. Рядом с Mountain даёт haste всем твоим." },
      { n: "5", card: "Reanimate", tag: "to board", text: "Существо из любого graveyard. Теряешь life равные его mana value." }
    ],
    sections: [
      {
        title: "Что делает командир",
        paragraphs: ["Mill не заменяет атаку. Return открывается только после combat damage, и существо входит tapped: в этот ход само оно не бьёт. Haste ему даёт Anger, если тот уже лежит в graveyard и на столе есть Mountain."],
        plays: [
          {
            line: "Supplier кладёт три карты ещё до командира и ещё три, когда умрёт. Каждый бой Terra mill две, бьёт игрока и за {2} возвращает существо с power 3 или меньше. Anger в graveyard даёт ему haste.",
            ticks: ["Supplier mill 3", "бой: mill 2", "урон игроку", "pay {2}", "входит tapped", "Anger: haste"],
            steps: [
              { card: "Stitcher's Supplier", via: "mill 3", text: "При выходе и при смерти. Graveyard наполняется раньше, чем Terra вообще атакует." },
              { card: "Terra, Herald of Hope", via: "pay {2}", text: "Каждый твой бой: mill 2 и flying. После урона игроку платишь {2}." },
              { card: "Anger", via: "haste", text: "Пока карта в graveyard и есть Mountain, у всех твоих существ haste. Вернутое бьёт сразу." }
            ]
          },
          {
            line: "Celes перекладывает руку в graveyard и берёт на одну карту больше. Когда Terra достаёт оттуда существо, Celes кладёт +1/+1 на всех твоих.",
            steps: [
              { card: "Celes, Rune Knight", via: "discard", bridge: "Сброшенные карты как раз то, что Terra поднимает после удара. Существо входит из graveyard, и Celes это видит.", text: "Discard сколько хочешь, draw на одну больше." },
              { card: "Terra, Herald of Hope", via: "+1/+1 всем", text: "Возвращает одно существо с power 3 или меньше. Celes раздаёт counters." }
            ]
          }
        ]
      },
      {
        title: "Подъём",
        paragraphs: ["Сначала мелочь через Terra, за {2} и с ограничением по power. Дорогое достают отдельные карты: за жизни, за позднюю ману или когда в graveyard уже толстая стопка."],
        beats: [
          { card: "Reanimate", stat: "life = MV", text: "Любое существо из любого graveyard. Пятёрка снимает пять жизней. Это не режим Terra." },
          { card: "Priest of Fell Rites", stat: "3 life", text: "Tap, 3 жизни, жертва: верни существо. Сам Priest встаёт из graveyard по unearth, с haste, и в конце хода изгоняется." },
          { card: "Stitch Together", stat: "7 карт", text: "За {B}{B} существо в руку. Если в graveyard уже семь карт, оно выходит сразу на стол." },
          { card: "Phoenix Down", stat: "MV ≤ 4", text: "Существо со стоимостью 4 или меньше выходит tapped. Шире, чем power 3 у Terra, но без атаки в этот ход." },
          { card: "Sun Titan", stat: "MV ≤ 3", text: "При выходе и при каждой атаке верни перманент со стоимостью 3 или меньше. Не только существо." },
          { card: "The Falcon, Airship Restored", stat: "{4}{B}", text: "Боевой урон и жертва Falcon: верни существо. Сам Falcon встаёт из graveyard за {4}{B}, tapped." }
        ]
      },
      {
        title: "Чем это заканчивается",
        paragraphs: ["Эти карты не прячутся. Их видно за ход до того, как стол соперника кончится."],
        beats: [
          { card: "Sepulchral Primordial", stat: "по 1 существу", text: "При выходе забери по одному существу из graveyard каждого соперника." },
          { card: "Rise of the Dark Realms", stat: "все существа", text: "Все существа из всех graveyard выходят под твоим контролем. Поздняя карта, не ход четыре." },
          { card: "Kefka, Dancing Mad", stat: "free cast", text: "В твой ход его indestructible. В конце хода изгони случайную карту из каждого чужого graveyard и можешь скастовать её бесплатно." },
          { card: "Ruinous Ultimatum", stat: "7 маны", text: "Destroy все неземли соперников. После него на их столе остаются земли." }
        ]
      },
      {
        title: "Пока колода добирается",
        beats: [
          { card: "Banon, the Returners' Leader", stat: "1 раз за ход", text: "Скастуй из graveyard существо, которое попало туда в этот ход не с поля. При атаке {1} и discard: draw." },
          { card: "Edgar, Master Machinist", stat: "артефакт tapped", text: "Раз за ход артефакт из graveyard, выходит tapped. При атаке +X/+0 по самому дорогому артефакту на столе." },
          { card: "Shadow, Mysterious Assassin", stat: "deathtouch", text: "Урон игроку и жертва другого перманента: две карты, соперники теряют жизни по стоимости жертвы." },
          { card: "Pitiless Plunderer", stat: "Treasure", text: "Другое твоё существо умерло. Treasure. Этой маной платят следующий подъём." },
          { card: "Morbid Opportunist", stat: "1 карта", text: "Кто-то кроме него умер. Одна карта, и только один раз за ход." }
        ]
      }
    ]
  },
  "Соня": {
    kicker: "60 карт · синий и зелёный · 20 жизней",
    hero: "Sokka, Lateral Strategist",
    level: {
      scale: "Kitchen",
      name: "Casual",
      note: "Стол собирается, но карты смотрят в разные стороны. Не тир Standard."
    },
    lead: "Сине-зелёные существа, в основном союзники. Выставляешь стол, атакуешь пачкой, чужое крупное существо бьёшь своим через fight.",
    stats: [
      { k: "Старт", v: "ход 1" },
      { k: "Sokka", v: "draw" },
      { k: "Suki", v: "= существа" },
      { k: "Fight", v: "×2" }
    ],
    turnsLine: "Существа стоят с первой маны. К четвёртому ходу Sokka атакует не один и за это берёт карту.",
    turns: [
      { n: "1", card: "Turtle-Duck", tag: "на стол", text: "Сразу на board. Позже за {3} power 4 и trample до конца хода." },
      { n: "2", card: "Dragon Sniper", tag: "deathtouch", text: "Тоже одна мана: reach, vigilance, deathtouch." },
      { n: "3", card: "Iguana Parrot", tag: "flying", text: "Flying. Каждое noncreature spell даёт ему +1/+1 до конца хода." },
      { n: "4", card: "Sokka, Lateral Strategist", tag: "draw", text: "Атакуй им и ещё кем-нибудь. Draw. Две копии." }
    ],
    sections: [
      {
        title: "Союзники, которые хотят соседей",
        paragraphs: ["Один союзник почти ничего не делает. Draw, token и рост включаются, когда их двое или трое и они атакуют вместе."],
        plays: [{
          line: "Sokka берёт карту, только если в эту атаку с ним пошёл ещё кто-то. Suki в той же атаке ставит tapped Ally 1/1, и её power равен числу твоих существ. Katara, если осталась tapped, кладёт на себя counter.",
          ticks: ["двое атакуют", "Sokka: draw", "Suki: Ally 1/1", "Katara: +1/+1"],
          steps: [
            { card: "Sokka, Lateral Strategist", via: "ещё один", text: "Vigilance. Атака вместе с другим существом: карта. Две копии." },
            { card: "Suki, Kyoshi Warrior", via: "Ally 1/1", text: "Power равен числу твоих существ. При атаке создаёт tapped атакующего Ally." },
            { card: "Katara, Bending Prodigy", via: "+1/+1", text: "В конце хода, если она tapped, кладёт на себя counter. Waterbend {6}: draw." }
          ]
        }],
        beats: [
          { card: "Ty Lee, Chi Blocker", stat: "flash", text: "При выходе tap цели, и она не untap, пока Ty Lee на столе." },
          { card: "Teo, Spirited Glider", stat: "draw, потом discard", text: "С flying атаковали: карта, потом discard. Discarded неземля кладёт +1/+1 на существ." },
          { card: "Earth Kingdom General", stat: "земля → 0/0", text: "При выходе земля становится существом 0/0 с haste и двумя +1/+1. Counters на существе можно обратить в жизни, раз за ход." },
          { card: "Turtle-Duck", stat: "{3} = 4", text: "Сидит с первой маны. За {3} до конца хода power 4 и trample, как раз в ту атаку, где Sokka хочет соседа." }
        ]
      },
      {
        title: "Если чужое существо мешает",
        paragraphs: ["Fight здесь считается от своих существ, и союзники делают его дешевле. Если бить нечем, существо проще выключить."],
        plays: [{
          line: "Rocky бьёт чужое существо силой твоего. Allies at Last кидает в одну цель сразу двоих и стоит на {1} дешевле за каждого Ally.",
          steps: [
            { card: "Rocky Rebuke", via: "fight", bridge: "Две копии. В цель идёт сила твоего существа, не фиксированный урон. Рядом Suki это уже не 2, а число существ на столе.", text: "Твоё существо наносит урон, равный своей силе." },
            { card: "Allies at Last", via: "−{1} / Ally", text: "Два твоих существа бьют одну цель. Дешевле на {1} за каждого Ally." }
          ]
        }],
        beats: [
          { card: "Retro-Mutation", stat: "Turtle 0/1", text: "Flash. Чужое существо становится Turtle 0/1 без способностей и не атакует." },
          { card: "Watery Grasp", stat: "не untap", text: "Существо не untap. Waterbend {5} замешивает его в библиотеку владельца." }
        ]
      },
      {
        title: "Крупнее, чем стартовая пачка",
        beats: [
          { card: "Walltop Sentries", stat: "×3", text: "Reach и deathtouch. Смерть при Lesson в graveyard: 2 жизни." },
          { card: "Primordial Pachyderm", stat: "×2", text: "Reach, trample, 2 жизни при выходе." },
          { card: "Invasion Tactics", stat: "+2/+2", text: "При выходе всем +2/+2 до конца хода. Урон Ally по игроку draw." },
          { card: "Long Feng, Grand Secretariat", stat: "+1/+1", text: "Другое существо или земля умерли. +1/+1 кому-то из твоих." }
        ]
      }
    ]
  }
};
