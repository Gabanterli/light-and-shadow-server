# World Continents and Regions Documentation

This document acts as the official registry and guide for all world continents, regions, and settlements in **Light and Shadow**.

## Continent Registry

### 1. Main Continent (Continente Principal)
* **Minimum Level:** 1 (Starter Zone 1-10, Midgame Progression 10-50)
* **Symbolic Monarch:** King Aldren Vaelor (Officially residing at the Royal Seat of Ravenshire)
* **Core Role:** Starter zone, open world sandbox progression, and global trade and maritime port.
* **World Boundaries:** Min X: 0, Max X: 1999, Min Y: 0, Max Y: 1999.
* **Geography & Biomes:**
  * **Plains:** 40% (Peaceful starting plains and farmland)
  * **Forests:** 30% (Dense woodlands sheltering wild beasts and timber)
  * **Hills:** 15% (Rolling highlands flanking the trade roads)
  * **Mountains:** 10% (Steep crags rich in mining resources)
  * **Coastal Zones:** 5% (Trade ports and harbors interfacing with other continents)
* **Political Structure & Factions:** Symbolic rule under King Aldren Vaelor from his royal court in Ravenshire, who maintains political balance between independent, self-governed cities:
  * *Human Kingdom Alliance:* Dominant feudal and aristocratic alliance.
  * *Merchant Guild Confederation:* Tycoons controlling trading routes and pricing.
  * *Military Order of Stone Tirith:* Elite defenders of law and border security.
  * *Independent Frontier Clans:* Outpost clans and hunters on the northern borders.
* **Economic Role:** Open sandbox economy ranging from starter weapons to complex dwarven crafting, formal guild commerce, political capital and noble affairs, global maritime trading, and resource harvesting.
* **Canonical Settlements & POIs:**
  1. **Ironhold Bastion (100,100):** Starter military bastion and safe hub. Includes beginner training grounds. Stationed with Mayor George and Guard Will. Establishes the starter economy.
  2. **Stone Tirith (500,500):** Grand Dwarven mountain city led by Thane Thromgar and Marshal Varian. Center of mining, forge economy, and weapon crafting.
  3. **Blackwater Bay (1200,1200):** Bustling coastal port and pirate hub led by Port Master Drake and Captain Redbeard. Hub of global shipping, smuggling, and maritime travel.
  4. **Ravenshire (1600,600):** Official commercial capital, official political capital, and royal seat of King Aldren Vaelor and Lady Genevieve. Sede of formal trade, diplomacy, high aristocracy, noble houses, guild contracts, and a grand marketplace economy (strictly non-military).
  5. **Thornwall (800,800):** Diplomatic mixed-race forest city of humans, elven descendants, and green orcs, led by Chieftain Garak and Elder Elenwe. Center of natural resources and forestry.

### 2. Nature Continent (Nature)
* **Minimum Level:** 1
* **Capital/Settlement:** Elarin
* **Description:** Lush forests, woodlands, and elven strongholds teeming with ancient forest guardians.

### 3. Shadow Continent (Shadow)
* **Minimum Level:** 1
* **Capital/Settlement:** Noctharyn
* **Description:** Dark, corrupted lands under an eternal twilight. Home to ancient necropoles and ashen wastes.

### 4. Fire Continent (Continente de Fogo)
* **Module Name:** `fire_continent`
* **Minimum Level:** 50
* **Capital/Settlement:** Pyra Magnus
* **Description:** A unified, highly dangerous volcanic world shaped by active magma rivers, basalt mountains, and industrial smithing. This continent acts as the primary global crafting hub and source of global fire energy.

#### 🌋 Geography & Zones:
* **Central Primordial Volcano:** (45% of continent) Core region, high-danger dungeon housing the Fire Demons.
* **Ash Plains:** (30% of continent) Mid-level resource farming and exploration area.
* **Forge Mountains:** (25% of continent) Rich iron ore ranges inhabited by the master-craftsman Fire Cyclopes.

#### 👺 Registered Races:
1. **Red Orcs:** Dominant surface warriors of Crimson Hollow with a tribal militaristic society and extreme heat resistance.
2. **Fire Cult Humans:** A highly religious human civilization that worships the Fire Primordial and is fully integrated into the political system.
3. **Fire Cyclopes:** Master blacksmiths of Molten Anvil responsible for crafting the absolute best weapons in the game.
4. **Fire Demons:** Extremely powerful entities that reside exclusively inside the central volcano dungeon (not part of surface civilization).

#### 👑 Political Structure:
* **Government Type:** Unified Arcane-Military Monarchy.
* **Ruler:** **Ignis Rex (Fire Mage King)**. He is possessed by a Lesser General of the Fire Primordial, who heavily influences all political decisions but does not completely override the King's identity.

#### 🏰 Major Cities:
1. **Pyra Magnus:** The capital city and political/arcane epicenter of the continent, hosting the grand palace of Ignis Rex.
2. **Crimson Hollow:** The Red Orc stronghold city, acting as the military expansion hub and war arena epicenter.
3. **Molten Anvil:** The Cyclops forge city, leveraging raw magma streams for the ultimate weapon and gear crafting systems.

### 5. Holy Continent (Sagrado)
* **Minimum Level:** 50
* **Capital/Settlement:** Luminaar
* **Description:** A divine territory guarded by paladins. Requires completion of the **Spiritual Trial (Provação Espiritual)** to unlock travel routes.

### 6. Ice Continent (Continente de Gelo)
* **Minimum Level:** 1
* **Capital/Settlement:** Elarisheim
* **Description:** A frozen wilderness of eternal snows and frost, where elven royalty and human settlers struggle against the severe climate.

---

## Hidden Subregions & Restricted Zones

### Ymirr's Hidden Cavern (Caverna de Ymirr)
* **Region ID:** `ymirr_hidden_cavern`
* **Parent Continent:** Ice Continent
* **Coordinates:** `[X: 5980, Y: 5965]`
* **Type:** Sacred Jötunn Cavern (Hidden Subregion)
* **Leader/Governor:** Ymirr Stonefrost
* **Status:** **UNDISCOVERED** and **RESTRICTED**

#### 🔒 Access Rules & Patch Details (`ice_hidden_region_patch`):
* **No Public Travel:** Removed completely from the public continent travel menu/tabs and world map UI.
* **No Direct Ship/Port Access:** Standard sea captains, port masters, and portal networks do not have direct routes to the cavern. Attempts to trigger direct coordinate-based transport requests (`CS_MOVE_REQUEST`) are blocked and rejected by the server movement filter.
* **Access Methods:** Can only be reached through special conditions:
  1. Future questlines (such as unlocking ancient runes).
  2. Traversal of the dangerous **Frozen Mountain Peaks** (Cumes Proibidos).
  3. Secret passage through the deep mines of **Khaz Tirith** (Dwarven dungeon/cavern access).

---

## Authoritative Cosmology: The Primordial System

The **Primordial System** serves as the core cosmological foundation of the world of **Light and Shadow**. All current and future systems, lore, quests, dialogue, and rituals must strictly adhere to these rules.

### 1. Nature of the Primordials
* **Conceptual Entities:** Primordials are not spawnable mobs, raid bosses, enemies, or directly fightable NPCs. They represent the absolute cosmic essence of major world domains.
* **Non-Interactive Gameplay:** Primordials exist outside normal gameplay systems (cannot be fought, killed, or cataloged in bestiaries). They are referenced strictly through lore, dialogues, rituals, quests, and world events.

### 2. Continental Primordial Mapping
Each major continent is bound to and influenced by exactly **one** Primordial entity:
* **Nature Continent:** Associated with the **Primordial of Nature / Life** (Emanation of wild growth and life force).
* **Ice Continent:** Associated with the **Primordial of Ice / Frost** (Sovereign of absolute zero and the glacial wilderness).
* **Fire Continent:** Associated with the **Primordial of Fire** (Imprisoned core of magma and cosmic flame).
* **Holy Continent:** Associated with the **Primordial of Light** (Divine source of eternal luminescence).
* **Shadow Continent:** Associated with the **Primordial of Shadow / Death** (The eternal twilight and ashen decay).
* **Abyssia:** Associated with the **Primordial of Abyss / Void** (The hungry dark and cosmic nothingness).

### 3. The Main Continent Exception (Neutral Equilibrium)
* **Strict Canonical Exception:** The **Main Continent (`main_continent`) has NO Primordial.**
* **Equilibrium Domain:** It exists completely outside direct Primordial dominion. This cosmological neutrality explains why multiple races (Humans, Dwarves, Elves, Green Orcs) coexist in harmony, why a delicate political balance is maintained, and why it became the central civilization and trading hub of the world.

### 4. Demon Creation Rule
* **Origin of Demonic Hierarchies:** Demons are not an independent or natural species. Every demonic entity and hierarchy is created directly by and originates from a **Primordial source** (e.g., Fire Demons are creations of the Primordial of Fire).
* **System Compliance:** Any future implementation of demonic creatures, rifts, or hierarchies must preserve this origin rule.

---

## Canonical Demon Bible: Primordial Creation Structure

The **Demon Bible** establishes the absolute authoritative rules governing all demonic systems, hierarchies, and ecology within the world of **Light and Shadow**.

### 1. Independent Demon Structure
* **Core Principle:** Each Primordial cosmic force creates and sustains its own independent **Demon Structure**.
* **Self-Contained Systems:** A Demon Structure is a fully self-contained hierarchical system operating purely under its parent Primordial's cosmic influence.

### 2. Precise Structural Hierarchy (Per Primordial)
Every active Primordial structure is composed of the following precise tiers:
* **3 Archegenerals:** Colossal generals of world-ending calamity who execute the direct cosmic will of the Primordial.
* **5 Commanders:** High commanders supervising legions and managing planetary incursions.
* **10 Minor Demons:** Named, highly lethal lieutenants governing localized areas, rifts, or dungeons.
* **Unlimited Common Demons (Bestiary Tier):** Standard grunts, foot-soldiers, and beasts that constitute the general monster population in associated zones.

Each of these tiers is unique, custom-designed, and exclusive to its parent Primordial.

### 3. Strict No Cross-Creation Rule
To maintain strict ecological and lore consistency, the following boundaries are absolute:
* **No Cross-Service:** Demons created by one Primordial **cannot** serve or pledge allegiance to another Primordial.
* **No Allegiance Switching:** Demonic entities are cosmologically bound to their creator; switching sides or defection to different primordial factions is impossible.
* **No Power Inheriting:** Demons cannot inherit, wield, or adapt power from other Primordials (e.g., a Fire Demon cannot absorb Frost/Ice primordial essences).
* **Isolated Hierarchies:** Demonic lineages cannot mix or cross-breed. Fire Demons belong solely to the Fire Primordial, Shadow Demons to the Shadow Primordial, and Abyss Demons to the Abyss/Void Primordial.

### 4. Game & System Design Compliance
* **Hunt-Driven Encounter System:** Demons are no longer continent-restricted or continent-locked. All creature types, including Demons and Dragons, may appear in any region, biome, dungeon, hunt, or encounter.
* **Encounter Overrides Ecology:** Spawn composition, density, difficulty, and encounter mix are determined dynamically by Hunt configurations, encounter design, narrative requirements, and world event logic, rather than rigid species-based geographic restrictions.
* **System Modularization:** All current and future backend databases, spawners, quests, and bestiaries treat Demon Structures as modular and independent systems keyed and organized by their parent **Primordial Source**.

---

## Canonical Dragon Bible: Light and Shadow

The **Dragon Bible** establishes the absolute authoritative rules and cosmological framework governing all dragons, draconic lineages, and hierarchies within the world of **Light and Shadow**.

### 1. Nature of Dragons
* **Ancient & Autonomous:** Dragons are an ancient, completely autonomous race that exists independently of the Primordials.
* **Distinct Entities:** They are **NOT** creations of Primordials, demons, magical constructs, or faction-controlled entities.
* **Sovereign agency:** Dragons possess full free will, deep intelligence, and self-directed existence.

### 2. Cosmological Position
The universal hierarchy is defined by distinct layers of consciousness and origin:
* **Primordials:** Conceptual, non-combatant cosmic forces representing the absolute essence of world domains.
* **Demons:** Manifestations and servant hierarchies created directly by and bound to Primordials.
* **Dragons:** An independent, ancient sovereign race existing entirely outside of Primordial creation.
* **Mortals:** The playable races, civilizations, and active guilds of the world.

### 3. Primary Dragon Lineages
There are exactly **three primary dragon lineages** in existence:
* **Fire Dragons** (Wielders of cosmic heat and magma affinity).
* **Ice Dragons** (Wielders of absolute frost and glacial cold affinity).
* **Shadow Dragons** (Wielders of umbral energy and twilight affinity).

Each of these lineages exists independently of any Primordial, arising from the ancient fabric of the universe itself.

### 4. Draconic Progression Hierarchy
Unlike demonic systems based on absolute obedience, draconic hierarchy is organized by natural progression based on age, accumulated power, and elemental essence:
1. **Lesser Dragons** (Drakes, wyverns, and hatchlings).
2. **Common Dragons** (Fully-grown, sentient adult dragons).
3. **Elder Dragons** (Ancient, long-lived dragons possessing vast intelligence).
4. **Ancient Dragons** (Legendary colossi spoken of only in oldest myths).
5. **Dragon Lords** (The highest-tier, supreme draconic rulers).

### 5. The Dragon Lord Supreme: Void
* **Apex Draconic Authority:** There exists one ultimate entity at the pinnacle of all draconic power: **Void, the Apex Dragon Lord**.
* **Identity Constraints:** **Void** is neither a Primordial nor a Demon. Void represents the highest draconic authority in existence, ruling with absolute sovereign will over draconic kind.

### 6. World Distribution and Spawn Rules
* **Not Region-Locked or Biome-Locked:** Dragons are not restricted or confined to any single continent, region, or biome. They may appear in any region, biome, dungeon, hunt, or encounter, determined by Hunt and Encounter configuration, narrative requirements, or world event logic.
* **Spawn Manifestations:** Dragons can appear as world bosses, dungeon bosses, raid encounters, apex hunt creatures, and special event spawns.
* **Elite Threat Gameplay:** Dragons function exclusively as powerful encounters, elite bosses, legendary global events, and high-risk exploration threats. They are **never** faction units or controlled armies of any kingdom.

### 7. The Core Ontological Distinctions
* **Dragons:** Governed by **Free Will**, ancient intelligence, and sovereign strength.
* **Demons:** Governed by **Absolute Obedience** to their specific Primordial creator.
* **Primordials:** Non-interactive, pure **Cosmic Concepts** and environmental anchors.

This fundamental ontological distinction must remain consistent across all gameplay systems, dialogues, spawning tables, and lore.

---

## Canonical PvE Combat Bible & Elemental Affinity System

This patch acts as the authoritative system design manual for the PvE Combat Model, resource management, and elemental class progression in **Light and Shadow**.

### 1. PvE Combat Philosophy
* **Tactical & Long-Duration:** Light and Shadow uses a tactical, slow, and attrition-based PvE combat model inspired by classic MMORPGs but modernized with active resource management.
* **Core Combat Pillars:**
  * Auto attack based on weapon attack speed.
  * Active skills and spells with low, active cooldowns.
  * Long Time-To-Kill (TTK).
  * Low monster density and high encounter significance.
  * Tactical resource management rather than mindless mob grinding.

### 2. Auto Attack System
Basic attacks are entirely automatic and governed strictly by weapon type attack speed:
* **Daggers:** 0.6 – 0.9 seconds
* **Swords:** 1.0 – 1.4 seconds
* **Bows:** 0.9 – 1.3 seconds
* **Staffs:** 1.2 – 1.8 seconds
* **Axes:** 1.6 – 2.2 seconds

Auto attacks form the fundamental backbone of sustained combat and physical resource generation.

### 3. Skills and Spells
While basic auto attacks sustain the rotation, skills and spells keep combat active via low cooldown tiers:
* **Quick Skills:** 1 – 3 seconds
* **Core Skills:** 4 – 8 seconds
* **Major Skills:** 10 – 25 seconds

Skills are rotated frequently during combat while avoiding excessive burst-damage gameplay.

### 4. Long Time-To-Kill (TTK)
Combat is deliberately slow, requiring players to focus on positioning, defensive execution, and resource sustain. Target TTK ranges are:
* **Common Monsters:** 40 seconds – 2 minutes
* **Elite Monsters:** 3 – 8 minutes
* **Mini Bosses:** 8 – 20 minutes
* **Bosses:** 20 – 45 minutes
* **Dragons / Demon Lords:** 30 – 90+ minutes

### 5. Low Monster Density
PvE encounters favor fewer, more tactical and threatening enemies (e.g., 2–4 dangerous monsters) rather than pulling large groups of trivial targets. This improves tactical depth, AI logic, server performance, and co-op gameplay.

### 6. Hybrid Mana System
Mana recovery is hybrid and primarily combat-driven:
$$\text{Mana Gain} = \text{Offensive Recovery} + \text{Defensive Recovery} + \text{Passive Regen}$$
* **Offensive Recovery:** Mana gained upon dealing damage.
* **Defensive Recovery:** Mana gained when successfully blocking, parrying, or mitigating incoming damage.
* **Passive Regen:** Very low outside combat and nearly negligible during active combat. This ensures active skill rotation generates the required mana resources.

### 7. Hybrid HP Sustain System
Sustain and survival depend on four layers: Potions, Healing Skills, Passive Regen, and Defensive Mechanics.
* **Anti-Potion Spam:** Potion usage is regulated by strict, separate cooldowns:
  * **Health Potion:** 8 – 15 seconds
  * **Mana Potion:** 12 – 20 seconds
  * **Rare Potions:** 30 – 60 seconds
* **Passive HP Regen:** Moderate outside combat, but minimal during combat.

### 8. Defensive Combat Mechanics
To reward mechanical skill expression, players have active defensive layers:
* **Block:** Reduces incoming physical damage.
* **Parry:** Mitigates damage and opens opportunities for counters.
* **Dodge:** Completely avoids incoming physical attacks.
* **Magic Barrier:** Absorbs incoming elemental/magical damage via barriers.

### 9. Elemental Combat & Status Effects
The combat system supports six elements affecting damage scaling, resistances, status effects, skill interactions, and boss mechanics:
* **Fire:** Inflicts *Burn* (damage over time).
* **Ice:** Inflicts *Slow / Freeze* (movement & speed mitigation).
* **Nature:** Inflicts *Poison / Root* (damage over time / crowd control).
* **Holy:** Inflicts *Purification / Anti-Undead* bonuses.
* **Shadow:** Inflicts *Corruption / Drain* (health sap).
* **Abyss:** Inflicts *Chaos / Corruption* (Note: Abyss exists only as a combat element and is **NOT** a playable affinity).

### 10. Elemental Player Affinity & Progression
Affinities are developed organically through gameplay, not through configuration menus.
* **Playable Affinities:** Fire, Ice, Nature, Holy, and Shadow.
* **Non-Playable:** Abyss.
* **Awakening Lock:** Each player can permanently awaken **only one** affinity.

### 11. Dual-Track Progression & Level System
Elemental affinity progresses independently from standard character levels. There is no traditional endgame ceiling:
* **Character Level Range:** 1 – 9999 (Long-term progression is a core pillar of the MMORPG; there is no hard level 200 ceiling).
* **Affinity Level:** 1 – 100 (Progress is stored independently for each playable affinity: Fire, Ice, Nature, Holy, and Shadow).

### 12. Affinity Progression Sources
Progression is achieved through a hybrid system with four sources:
1. **Skill Usage:** Actively utilizing elemental skills in combat grants elemental XP (No safe-zone abuse allowed; must be valid combat actions).
2. **Monster Kills:** Slaying creatures associated with specific elements (e.g., Fire demons, Fire cyclopes, Fire dragons) grants respective affinity XP.
3. **Regional Exposure:** Passive affinity growth when actively fighting on corresponding elemental continents (e.g., Fire Continent gives Fire affinity bonus).
4. **Elemental Trials:** Completing major elemental-themed quests or specialized rituals.

### 13. Elemental Awakening
Awakening occurs permanently and irreversibly only when:
$$\text{Character Level} \ge 100 \quad \text{AND} \quad \text{Dominant Affinity Level} = 100$$
Upon awakening, the chosen affinity becomes permanently and irrevocably locked into the `awakened_affinity` data model. Secondary awakenings or changes to `awakened_affinity` are strictly blocked.
The base class permanently evolves into an elemental subclass:
* **Knight** $\rightarrow$ *Fire Knight*, *Ice Knight*, etc.
* **Mage** $\rightarrow$ *Shadow Mage*, etc.
* **Cleric** $\rightarrow$ *Holy Cleric*, etc.

### 14. Elemental Gear Compatibility
* **No Equipment Lock:** Players may equip gear from affinities different from their own; mismatching does not block weapon/armor use.
* **Canonical Efficiency Rates:**
  * **Same Affinity:** 100% gear efficiency.
  * **Neutral Affinity:** 75% gear efficiency (standard for non-opposing differences).
  * **Opposed Affinity:** 50% gear efficiency (severe stats penalty).
* **Elemental Opposition Rules:**
  * **Fire (Fogo) ↔ Ice (Gelo):** Direct opposed pair.
  * **Holy (Sagrado) ↔ Shadow (Sombra):** Direct opposed pair.
  * **Nature (Natural):** Nature has no direct opposed affinity. Nature equipment interacting with non-matching affinities defaults to Neutral affinity efficiency (75%).

This preserves build diversity while fully rewarding deep specialization.

### 15. Canonical Experience Curve & Combat Balancing

To prevent progression stagnation while avoiding the need for continuous database tier expansions, **Light and Shadow** employs a non-linear, frontloaded experience scaling system.

#### 15.1 Experience Curve Phases
1. **Early Game (Levels 1–100) — FAST:**
   * **Purpose:** Onboarding, class choice (Vocation) at Level 10, region discovery, and unlocking the permanent Elemental Awakening at Level 100.
   * **Balancing:** XP requirements grow smoothly ($XP_{\text{needed}} = \text{Level}^2 \times 100$). Slaying same-level mobs and completing local quests provides fast leveling to avoid tedious grinding before Awakening.
2. **Mid Game (Levels 101–200) — MODERATE:**
   * **Purpose:** Subclass specialization, elemental optimization, and active dungeon hunting.
   * **Balancing:** Slowdown begins. XP requirements scale with a progressive multiplier to create a noticeable transition ($XP_{\text{needed}} = \text{Level}^2 \times 100 \times [1 + (\text{Level}-100) \times 0.05]$).
3. **Advanced Game (Levels 201–400) — SLOW:**
   * **Purpose:** Build optimization, high-level dungeon farming, elite hunting, and prestige progression.
   * **Balancing:** Significant grind. Each level represents a prestigious achievement ($XP_{\text{needed}} = \text{Level}^2 \times 600 \times [1 + (\text{Level}-200) \times 0.5]$).
4. **Endurance Game (Levels 401–9999) — VERY SLOW:**
   * **Purpose:** Pure status recognition, endless growth fantasy, and competitive rankings.
   * **Balancing:** Hyper-exponentials are applied. Levels are extremely rare. Standard progression shifts entirely away from raw levels to gear, collection completion, and boss mastery.

#### 15.2 System Reward Synchronization
* **Monster Kill XP:** Scaled with sharp diminishing returns to prevent macro-botting at high levels.
  * *Levels 1-100:* $10 \times \text{Level}$ XP.
  * *Levels 101-200:* $1,000 + (\text{Level}-100) \times 15$ XP.
  * *Levels 201-400:* $2,500 + (\text{Level}-200) \times 2.5$ XP.
  * *Levels 401+:* Diminishing return cap of $5,000$ XP max.
* **Quest XP Rewards:** Designed to reward active play while preventing leveling inflation.
  * *Levels 1-100:* $200 \times \text{Level}$ XP.
  * *Levels 101-200:* $20,000 + (\text{Level}-100) \times 300$ XP.
  * *Levels 201-400:* $50,000 + (\text{Level}-200) \times 125$ XP.
  * *Levels 401+:* Hard cap of $100,000$ XP.
* **Dungeon Completion XP:** Decreases as a percentage of the total level pool to maintain endgame longevity.
  * *Levels 1-100:* $5\% \text{ to } 15\%$ of next level per clear.
  * *Levels 101-200:* $0.2\% \text{ to } 5\%$ of next level per clear.
  * *Levels 201-400:* $< 0.1\%$ of next level.
  * *Levels 401+:* Infinitesimal ($< 0.0001\%$).

#### 15.3 Combat Balancing Implications
Raw character level alone must **not** dominate combat performance. As a player transitions into the Mid Game and beyond, the sources of character power shift dramatically:

| Progression Phase | Level Base Stats | Gear & Synergy | Skill Execution |
| :--- | :---: | :---: | :---: |
| **Early Game (1-100)** | 70% | 10% | 20% |
| **Mid Game (100-200)** | 40% | 30% | 30% |
| **Advanced Game (200-400)** | 15% | 50% | 35% |
| **Endurance Game (400+)** | 2% | 60% | 38% |

This design ensures that high-level players are celebrated for their dedication, but cannot trivialize endgame bosses or PvP combat simply by out-leveling standard limits. True power lies in gear builds, elemental synergy, and perfect action execution.

### 15.4 Cooldown Reduction (CDR) System & Restrictions

Cooldown Reduction (CDR) is allowed, but it is a highly restricted modifier. It is NOT treated as a common stat.

<yaml id="cdr-lock-rule">
cooldown-reduction-rule:
  text: "Cooldown Reduction is an ultra-rare utility modifier restricted mainly to Tier 4 and Tier 5 items, with a global maximum cap of 15%."
</yaml>

Key canonical rules for the CDR system:
* **Ultra-Rare Modifier:** CDR is highly limited and does not appear on typical gear or low-tier drops.
* **Tier Restriction:** Primarily appears in Tier 4 and Tier 5 items. Tier 1–3 items normally do not contain CDR.
* **Low-Impact Design:** CDR must always remain low-impact to safeguard the high-TTK tactical combat philosophy of the game.
* **Global Hard Cap:** The maximum total CDR from all sources combined (including equipment, active/passive skills, affixes, offhands, temporary buffs, consumables, and any other sub-systems) is strictly capped at **15%**. Any CDR beyond this cap is completely ignored.

### 15.5 HP / Mana Level Scaling System

Base HP and Base Mana defined by the classes are Level 1 baseline values. As characters level up, these resources scale upward using vocation-specific curves with an infinite level soft-cap system to prevent uncontrolled power creep.

<yaml id="level-scaling-rule">
level-scaling-rule:
  text: "Base HP and Base Mana represent class starting values and scale upward with level progression using vocation-specific scaling curves."
</yaml>

<yaml id="final-resource-formula">
resource-scaling-formula:
  text: "Final HP and Mana are calculated from base class values, level scaling, and equipment bonuses."
</yaml>

<yaml id="infinite-level-softcap-rule">
infinite-level-softcap-rule:
  text: "Character level progression is infinite but follows soft-cap scaling to prevent uncontrolled power creep."
</yaml>

Key rules for resource scaling:
* **Vocation-Specific Scaling:** Level progression grants HP and Mana dynamically based on the character's vocation. For instance, Knights gain massive HP and minimal Mana per level, whereas Mages gain massive Mana and minimal HP.
* **Structural Separation:** Growth from levels is handled as a separate layer ("Level Gain") and does not alter the immutable starting Base HP and Base Mana values.
* **Calculation Layering:** The final totals are calculated by layering Base Value, Level Scaling, and Equipment Bonuses deterministically.
* **Infinite Level Soft-Cap Bands:** While level progression is technically infinite (Tibia-like model), the HP and Mana growth per level decreases over progressive bands:
  * **Band 1 (Levels 1–200):** 100% scaling efficiency.
  * **Band 2 (Levels 201–500):** 50% scaling efficiency.
  * **Band 3 (Levels 501+):** 25% scaling efficiency.

As an example, Knights gain +15 HP and +2 Mana per level in Band 1, +7.5 HP and +1 Mana in Band 2, and +3.75 HP and +0.5 Mana in Band 3.

### 15.6 Elemental Progression Cap

Elemental Affinity allows players to specialize in specific paths of magic and defense. However, to maintain tactical balance, these progressions are strictly capped.

<yaml id="elemental-cap-rule">
elemental-cap-rule:
  text: "Elemental affinity progression has a hard maximum level of 100."
</yaml>

Key rules for elemental progression:
* **Hard Maximum Cap:** Elemental affinity has a strict hard maximum cap of level 100. No progression beyond level 100 is allowed.
* **Mastery Definition:** Level 100 affinity represents maximum possible elemental mastery.
* **Applicable Affinities:** This hard cap applies universally to all elemental categories: Fire, Ice, Holy, Shadow, and Nature.

### 15.7 Character Power Hierarchy

To organize and structure how character stats are determined, the game adheres to a strict four-layer power structure.

<yaml id="power-hierarchy-rule">
power-hierarchy-rule:
  text: "Character power is structured through four layers: Level (baseline growth), Skill (mastery growth), Gear (main power source), and Elemental specialization (capped at level 100)."
</yaml>

The four layers of character power are defined as:
1. **Level:** Infinite progression with soft-cap scaling (baseline growth layer).
2. **Skill:** Infinite progression with diminishing effective combat contribution (mastery layer).
3. **Gear:** The main, primary source of power for characters, allowing highly specialized gear builds (equipment layer).
4. **Elemental:** Specialization system capped at a hard maximum of level 100 (synergy layer).


### 16. Canonical Monster AI Bible — Light and Shadow

#### 16.1 Global AI Philosophy
All hostile entities operate using **Semi-Intelligent AI (AI Complexity Tier B)**.
> **Canonical Rule:** *“Monsters behave according to species instincts, combat archetype, and environmental context.”*

Monster behavior is determined by:
* **Species**
* **Combat Archetype**
* **Rank**
* **Elemental Alignment**
* **Hierarchy** (especially if demon or dragon)
* **Territory**

Monsters do not behave like simplistic scripted targets, but also do not behave like highly competitive human players.

#### 16.2 Universal AI Archetypes
Every hostile creature belongs to one of six canonical AI archetypes:

1. **Bruiser / Tank**
   * *Role:* Frontline pressure, absorbing damage, blocking player movement.
   * *Behavior:* Aggressively closes distance, prefers melee combat, protects fragile nearby allies.
   * *Traits:* High HP, high defense, low mobility.
2. **Predator / Assassin**
   * *Role:* Target and kill vulnerable targets.
   * *Behavior:* Prioritizes weak targets, deals burst damage, utilizes rapid repositioning.
   * *Target Priority:* (1) Low HP targets, (2) Low armor targets, (3) Ranged/caster units.
3. **Ranged Hunter**
   * *Role:* Ranged pressure.
   * *Behavior:* Kiting, distance maintenance, continuous repositioning.
   * *Preferred Combat Range:* 6–12 tiles.
4. **Caster**
   * *Role:* Magical pressure.
   * *Behavior:* Core spell rotations, Area-of-Effect (AoE) damage, casting debuffs.
   * *Priority:* Clustered groups, low magic resistance targets, stationary players.
5. **Support**
   * *Role:* Sustain and empower allies.
   * *Behavior:* Healing, shielding, and buffing targets.
   * *Priority:* (1) Commander, (2) Frontline Bruiser, (3) Nearest ally.
6. **Elite Commander**
   * *Role:* Battlefield control.
   * *Behavior:* Coordinates encounters, applies support/combat auras, alters combat flow.
   * *Aura Examples:* Attack speed boost, defense boost, frenzy, fear.

#### 16.3 Aggro System
The aggro system uses a **Hybrid Threat** system.
> **Canonical Rule:** *“Monster aggro is determined by both generated threat and archetype-specific instincts.”*

Aggro is calculated via two distinct layers:
1. **Threat Value:**
   Generated actively through:
   * Damage Dealt (100 damage = 100 threat baseline)
   * Healing Conceded (100 healing = 50 threat baseline)
   * Taunt skills & support actions
2. **Instinct Modifier:**
   Archetypes modify target preference ratios:
   * *Bruiser:* 80% threat / 20% instinct
   * *Predator:* 40% threat / 60% vulnerability (low HP/armor)
   * *Ranged Hunter:* Prioritizes range efficiency
   * *Caster:* Prioritizes clustered targets
   * *Support:* Prioritizes ally sustain
   * *Elite Commander:* Uses hybrid decision logic

*Note: Tank classes must NOT automatically maintain aggro without proper threat generation and positioning.*

#### 16.4 Leash System
Leash behaviors are managed through a **Dynamic Leash** system.
> **Canonical Rule:** *“Every hostile entity has a species-based leash behavior determined by aggression, intelligence, rank, and territorial instinct.”*

Leash Tiers:
* **Tier 1 — Passive Territorial (6–12 tiles):** Goblins, weak beasts.
* **Tier 2 — Standard Aggressive (15–30 tiles):** Wolves, orcs, skeletons.
* **Tier 3 — Hunter Pursuit (40–80 tiles):** Predators, assassins, elite beasts.
* **Tier 4 — Commander Territory (100+ tiles):** Commanders, minor demons.
* **Tier 5 — Apex / Legendary (No conventional leash):** Dragons, archegenerals, legendary bosses.

Aggro reset occurs immediately upon: (1) no valid target in range, (2) hard leash break, or (3) exploit detection.

#### 16.5 Pack Aggro
Pack aggro utilizes a **Solo Aggro** philosophy.
> **Canonical Rule:** *“Aggro is primarily individual.”*

Rules:
* Attacking one monster does **NOT** automatically aggro nearby monsters.
* Nearby hostile entities remain neutral unless independently triggered.
* *Exceptions allowed ONLY through explicit boss mechanics:* Summoned adds, reinforcement skills, or scripted boss events.
* Passive social aggro does not exist. This supports controlled pulls, fewer simultaneous enemies, and tactical combat readability.

#### 16.6 Pathfinding
Pathfinding uses **Smart Pathing**.
> **Canonical Rule:** *“Hostile entities use obstacle-aware navigation.”*

* **Capabilities:** Obstacle avoidance, route recalculation, shortest reachable route selection.
* **Obstacles Include:** Trees, rocks, buildings, ruins, walls.
* **Anti-Exploit Detection:** Stuck state detection, unreachable target detection, path failure timeout.
* *AI is smart but not omniscient; monsters do not predict movement like competitive players.*

#### 16.7 Survival / Flee Behavior
Survival behavior is species dependent.
> **Canonical Rule:** *“Survival behavior depends on species psychology, intelligence, hierarchy, and instinct.”*

Four survival modes are observed:
1. **Cowardly:** Flees at 15–30% HP (e.g., goblins, scavengers, weak bandits).
2. **Tactical Retreat:** Temporarily retreats and repositions (e.g., hunters, assassins, archers).
3. **Frenzy Under Death:** At low HP, gains attack speed and damage but loses defense (e.g., beasts, berserkers, red orcs).
4. **Death Before Retreat:** Never flees (e.g., demons, undead, zealots).
*SPECIAL DEMON RULE: Demons never flee unless explicitly ordered by superior hierarchy.*

#### 16.8 Dragon Survival Law
> **Canonical Rule:** *“A dragon never retreats once true combat has begun.”*

Dragons:
* Never flee
* Never surrender
* Never retreat
* Combat outcome versus a dragon must always resolve as: **player dies** OR **dragon dies**.
* At low HP, dragons enter **Draconic Rage** (attack speed increase, breath frequency increase, spell intensity increase).
* *Applies to:* Lesser dragons, elder dragons, ancient dragons, dragon lords (Void Dragon Lord follows this law absolutely).

#### 16.9 Boss Phase Intelligence
Boss AI uses a **Phase-Based Boss AI** system.
> **Canonical Rule:** *“Boss behavior changes according to predefined health thresholds.”*

Health Thresholds:
* **Phase 1 — Opening (100–76% HP):** Base behavior.
* **Phase 2 — Escalation (75–51% HP):** New abilities unlock.
* **Phase 3 — Dangerous (50–26% HP):** Severe mechanics activate.
* **Phase 4 — Final / Enrage (25–0% HP):** Maximum aggression.

Phase Transition Types:
1. **Berserker Phase:** Higher damage.
2. **Summoner Phase:** Summons adds.
3. **Control Phase:** Debuffs / crowd control.
4. **Transformation Phase:** Visual and moveset transformation.


# 17. CANONICAL MONSTER BESTIARY BIBLE — LIGHT AND SHADOW

## SECTION 1 — BESTIARY CLASSIFICATION SYSTEM

Canonical rule:
> *“All hostile entities in Light and Shadow are classified by Bestiary Families and Subfamilies rather than artificial rarity ranks.”*

“Bestiary Families classify biological and ontological categories only.”

Families define:
* biological nature
* ontological category
* ecological role
* broad environmental distribution

Families DO NOT define:
* intelligence tier
* tactical sophistication
* combat behavior
* threat prioritization
* AI instinct weights

These properties must be defined independently by Monster AI systems.

**Canonical Rule: Family ≠ Intelligence Tier**

Examples:
* **Humanoids** can contain primitive goblins, tactical orcs, or highly intelligent cultists.
* **Dragons** can contain instinct-driven juveniles, highly strategic elder dragons, or sovereign Dragon Lords.
* **Demons** can contain mindless lesser manifestations, highly strategic Commanders, or catastrophic Archegenerals.

Combat intelligence and tactics must reference Monster AI profiles instead.

Monster classification hierarchy:
```
Family
└── Subfamily
    └── Species
```

Artificial rarity labels such as:
* Common
* Rare
* Epic
* Legendary

must **NOT** be used as official bestiary classification.
These labels may exist internally for balancing or encounter design but are not canonical lore classifications.

## SECTION 2 — OFFICIAL BESTIARY FAMILIES

Light and Shadow has exactly 20 official Bestiary Families.

1. **Humanoids**  
   *Examples:* goblins, orcs, bandits, pirates, cultists
2. **Beasts**  
   *Examples:* wolves, bears, boars
3. **Predators**  
   *Examples:* dire wolves, panthers, sabertooths
4. **Insects / Vermin**  
   *Examples:* giant ants, spiders, scorpions, beetles
5. **Reptilians**  
   *Examples:* lizardmen, basilisks, serpents, nagas
6. **Avians**  
   *Examples:* harpies, vultures, ravens, giant eagles
7. **Aquatics**  
   *Examples:* sharks, sirens, krakens, sea serpents
8. **Amphibians**  
   *Examples:* giant frogs, toads, salamanders
9. **Undead**  
   *Examples:* skeletons, zombies, liches, wraiths
10. **Spirits**  
    *Examples:* ghosts, wisps, phantoms
11. **Elementals**  
    *Examples:* fire elementals, ice elementals, storm elementals
12. **Flora Creatures**  
    *Examples:* treants, carnivorous plants, vine horrors
13. **Fungi**  
    *Examples:* mushroom walkers, plague spores, mycelial horrors
14. **Constructs**  
    *Examples:* golems, animated armor, arcane sentinels
15. **Cursed Beings**  
    *Examples:* werewolves, cursed knights, corrupted monks
16. **Aberrations**  
    *Examples:* void horrors, flesh mutants, anomalies
17. **Demons**  
    *Description:* Special canonical family tied to Primordials.
18. **Dragons**  
    *Description:* Special canonical family independent from Primordials.
19. **Titans / Colossals**  
    *Examples:* giants, cyclops, colossi
20. **Celestials**  
    *Examples:* seraphs, sacred guardians, divine beasts

## SECTION 3 — HYBRID SPAWN SYSTEM

Canonical rule:
> *“Respawn behavior depends on species category, ecological role, and encounter importance.”*

Spawn system uses **Hybrid Spawn**.

Four canonical spawn categories:

1. **Standard Spawn**  
   *Description:* Respawn by fixed timer. Used for common farm monsters.  
   *Typical respawn:* 30 seconds → 5 minutes  
   *Examples:* goblins, wolves, skeletons, spiders  
2. **Dynamic Ecosystem Spawn**  
   *Description:* Respawn depends on player density, hunting pressure, and alive population. Used for ecological creatures.  
   *Examples:* beasts, predators, flora, aquatics  
3. **Rare Spawn**  
   *Description:* Semi-random spawn windows. Characteristics: long timers, multiple locations, low frequency.  
   *Typical respawn:* 30 minutes → 24 hours+  
4. **Legendary Spawn**  
   *Description:* No simple respawn timer. Triggered by world events, hidden conditions, quest progression, or special world states.

## SECTION 4 — DEMON SPAWN LAW

Canonical rule:
> *“Demonic manifestation depends on hierarchy level, corruption intensity, and Primordial influence.”*

Demons use **Hybrid Demon Spawn**.

1. **Common Demons**  
   *Description:* Spawn fixed in corrupted areas.  
   *Examples:* volcanic demon zones, corrupted dungeons, deep shadow territories  
2. **Minor Demons**  
   *Description:* Semi-fixed spawn. These are elite hunt mobs and never bosses.  
   *Respawn:* 30 minutes → 12 hours  
3. **Commanders**  
   *Description:* Do not naturally respawn. Manifestation requires: rituals, high corruption, dungeon events, summoning conditions.  
4. **Archegenerals**  
   *Description:* Never behave as ordinary open-world mobs. Manifest only through major raid events, storyline catastrophes, or dimensional breaches.

Hierarchical Summoning Rule:
> *“Higher demons may manifest lower demons. Lower demons cannot summon higher hierarchy.”*

*Examples:*
* Commander may summon Common Demons
* Archegeneral may manifest Commanders
* Minor Demon cannot manifest Commanders

Primordial Influence Rule:
> Each Primordial continent increases activity of demons aligned with that Primordial.

Main Continent Rule:
> The Main Continent has no Primordial. However, in accordance with the Hunt-Driven Encounter System, demons of any primordial alignment may spawn in any region, biome, dungeon, hunt, or encounter on the Main Continent based on Hunt configurations, encounter design, narrative requirements, or world events.

## SECTION 5 — DRAGON SPAWN LAW

Canonical rule:
> *“Dragons inhabit the world according to territorial instinct, age, elemental affinity, and individual will.”*

Dragons use **Hybrid Dragon Spawn**.

1. **Nest Spawn**  
   *Description:* Used by young dragons.  
   *Examples:* lava caves, frozen caverns, shadow ruins  
   *Applies to:* lesser dragons, juvenile dragons  
2. **Lair Spawn**  
   *Description:* Used by mature dragons.  
   *Canonical rule:* *“One lair equals one apex dragon territory.”*  
   *Examples:* volcanoes, mountain peaks, glaciers, ancient ruins  
   *Applies to:* adult dragons, elder dragons  
3. **Roaming Apex Spawn**  
   *Description:* Used by powerful dragons capable of migrating. Characteristics: global movement, unpredictable location, rare sightings.  
   *Applies to:* ancient dragons, elder shadow dragons  
4. **Sovereign Spawn**  
   *Description:* Reserved for apex dragons.  
   *Examples:* Dragon Lords, Void Dragon Lord  
   *Rules:* no traditional respawn, regional influence, server-scale encounters  

Territorial Law:
> *“Two apex dragons rarely tolerate the same territory.”*

*Consequences:*
* Dragon Lords avoid coexistence
* Ancient Dragons may fight over territory
* Lesser Dragons may inhabit dominated territory

Dragon Hoarding Law:
> *“Dragons accumulate valuable objects within their lairs.”*

This rule is purely lore-level at this stage. Systems of loot generation, market values, rarity logic, drop tables, or reward formulas remain completely unimplemented until the Loot System Bible is approved.

Allowed canonical interpretation: Dragon lairs may contain gold, artifacts, relics, ancient objects, and rare resources as environmental narrative elements. No drop probabilities, market values, rarity tables, or reward formulas may be defined yet.


==================================================================
CANONICAL LOOT SYSTEM BIBLE — LIGHT AND SHADOW
==============================================

SECTION 1 — CORE LOOT PHILOSOPHY

Light and Shadow follows the loot philosophy:
B+ — Sparse but Meaningful Loot

Canonical rule:
> *“Loot should remain scarce enough to preserve economic value, but meaningful enough to justify long-duration combat encounters.”*

The game does NOT follow loot-flood MMORPG design.
Combat is long-duration with:
* high Time-To-Kill
* low monster density per hunt
* meaningful individual encounters
Therefore loot must feel valuable.

==================================================================
SECTION 2 — DROP FREQUENCY
==========================

Canonical rule:
> *“On average, 50–60% of monster kills should generate meaningful material rewards.”*

Meaningful reward includes:
* currency
* crafting materials
* monster parts
* reagents
* essences
* consumables

This does NOT imply equipment drops.

Drop frequency by encounter type:
* Common Monsters: 45–60%
* Elite Monsters: 70–90%
* Bosses: 100%
* Dragons: 100%

==================================================================
SECTION 3 — LOOT LAYERS
=======================

Loot is divided into four layers.

1. Guaranteed Reward Layer
   Examples: gold, materials, components
2. Common Loot Layer
   Examples: bones, pelts, monster parts, essences
3. Rare Loot Layer
   Examples: runes, gems, recipes, enchant materials
4. Exceptional Loot Layer
   Examples: relics, artifacts, apex rewards

==================================================================
SECTION 4 — DEMONIC LOOT LAW
============================

Canonical rule:
> *“All demonic entities may drop demonic-exclusive materials, with loot quality scaling by hierarchy.”*

Hierarchy behavior:
* **Common Demons:**
  * Drop: demonic ash, corrupted blood, low demonic essence
* **Minor Demons:** (Minor Demons are elite hunt mobs, NOT boss encounters)
  * Drop: demon cores (low grade), infernal fragments, corrupted crystals
* **Commanders:**
  * Drop: high demon cores, command sigils, primordial shards
* **Archegenerals:**
  * Drop: primordial relics, abyssal artifacts, apex catalysts

Canonical loot philosophy for demons:
> *“Demonic loot favors crafting and ritual materials over equipment drops.”*

==================================================================
SECTION 5 — EQUIPMENT ECONOMY
=============================

Light and Shadow uses Hybrid Equipment Economy.

Canonical ratio:
* 40% direct monster drops
* 60% player-crafted equipment

Canonical rule:
> *“Equipment enters the economy through both direct monster drops and player-driven crafting.”*

Crafting remains the dominant economic force.

==================================================================
SECTION 6 — FIXED ITEM IDENTITY LAW
===================================

Canonical rule:
> *“Items of the same type are always identical in stats, requirements, and properties.”*

Example: Every Iron Sword is identical.

The following systems are explicitly forbidden:
* random affixes
* item quality rolls
* stat variance
* procedural item generation
* randomized equipment stats
* evolving equipment

Progression happens via equipment replacement, not item mutation.

==================================================================
SECTION 7 — CURRENCY SYSTEM
===========================

Light and Shadow uses a four-tier currency hierarchy.

Official currencies:
1. Bronze Coin
2. Silver Coin
3. Gold Coin
4. Diamond Coin

Canonical conversion rates:
* 100 Bronze = 1 Silver
* 100 Silver = 1 Gold
* 100 Gold = 1 Diamond Coin

Equivalent values:
* 1 Gold = 10,000 Bronze
* 1 Diamond Coin = 1,000,000 Bronze

Currency philosophy:
* **Bronze:** mass circulation, early game
* **Silver:** mid economy, services
* **Gold:** late-game trade, crafting
* **Diamond:** ultra-high-value economy, rare trades, guild-scale transactions

==================================================================
SECTION 8 — CURRENCY SOURCES
============================

Currency enters the economy through hybrid sources:
1. Monster Drops
2. Vendor Selling
3. Quests / Contracts
4. Player Market

Canonical rule:
> *“Biological creatures rarely carry currency directly, while civilized intelligent creatures do so more frequently.”*

Frequent currency carriers: bandits, pirates, humanoids, mercenaries, cultists
Rare currency carriers: beasts, dragons, flora, insects, elementals

==================================================================
SECTION 9 — LOOT RETRIEVAL SYSTEM
=================================

Light and Shadow uses:
Direct Inventory Auto-Loot with Custom Loot Filter

Canonical rule:
> *“Loot is transferred directly to player inventory upon monster death, subject to player-defined filter rules.”*

Pipeline:
Monster Death → Loot Roll → Filter Validation → Inventory Capacity Check → Auto Transfer OR Corpse Storage

==================================================================
SECTION 10 — LOOT FILTER SYSTEM
===============================

Players may configure loot filters by:
1. Item Type (Examples: gold, materials, consumables)
2. Item Name (Examples: Demon Core, Dragon Scale)
3. Value Threshold (Example: Ignore loot below vendor value threshold)

==================================================================
SECTION 11 — INVENTORY OVERFLOW LAW
===================================

Canonical rule:
> *“If player inventory cannot receive eligible loot, rejected items remain inside the monster corpse.”*

Overflow loot is stored in corpse.
Monster corpses become manually lootable only when overflow exists.

==================================================================
SECTION 12 — CORPSE PERSISTENCE
===============================

Loot corpses persist:
10–15 minutes (Default: 12 minutes)

Corpse states:
1. **Empty Corpse:** Short despawn
2. **Loot Corpse:** Persists while loot remains
3. **Protected Corpse:** Boss / special encounter logic


==================================================================
CANONICAL DEATH PENALTY BIBLE — LIGHT AND SHADOW
================================================

SECTION 13 — HARDCORE DEATH PHILOSOPHY

Light and Shadow uses Hardcore PvE Death.

Canonical rule:
> *“Death in PvE must carry meaningful consequences.”*

Death creates: risk, tension, preparation value.

==================================================================
SECTION 14 — ITEM LOSS LAW
==========================

Canonical rule:
> *“Upon PvE death, players may lose both equipped items and backpack inventory.”*

Potential losses include:
* **Equipped:** weapon, armor, accessories, shield
* **Inventory:** loot, materials, consumables, carried currency

==================================================================
SECTION 15 — BLESSING SYSTEM
============================

Blessings already exist in the game as a full-risk insurance layer and major gold sink. There are exactly 7 Blessings in total.

Canonical rule:
> *“Blessing protection is linear. 7 active blessings grant full 100% protection against PvE item loss.”*

### Blessing Protection Formula
```
Final Death Loss = Base Death Penalty * (1 - Active Blessings / 7)
```

### Blessing Protection Table
* **0 Blessings:** 0% (Full loss)
* **1 Blessing:** 14.285% protection
* **2 Blessings:** 28.571% protection
* **3 Blessings:** 42.857% protection
* **4 Blessings:** 57.142% protection
* **5 Blessings:** 71.428% protection
* **6 Blessings:** 85.714% protection
* **7 Blessings:** 100% protection (No loss of equipped gear, backpack items, carried loot, or carried currency)

### Blessing Consumption Law
> *“Any PvE death removes ALL active blessings simultaneously.”*

Regardless of how many blessings were active (even if fully protected at 7), a single PvE death strips the player of all active blessings, resetting them to 0 blessings.

==================================================================
SECTION 16 — PLAYER CORPSE LOOT LAW
===================================

Light and Shadow uses Open Loot for player corpses.

Canonical rule:
> *“Items dropped upon player death become accessible to any player capable of reaching the corpse.”*

Open loot applies to: owner, allies, strangers, opportunistic looters.
Corpse recovery becomes emergent gameplay.

==================================================================
SECTION 17 — BOSS REWARD PHILOSOPHY
===================================

Bosses use Personal Loot.

Canonical rule:
> *“Boss rewards are generated individually for each eligible participant.”*

There is:
* no shared loot pool
* no floor loot competition
* no first-click loot

Each eligible player receives individual rewards.

==================================================================
SECTION 18 — BOSS ELIGIBILITY SYSTEM
====================================

Boss rewards require Hybrid Eligibility Validation.

Canonical rule:
> *“Players must satisfy both participation duration and contribution thresholds.”*

Valid contribution includes: damage, healing, tanking, support actions.

Anti-exploit goals:
* prevent AFK leeching
* prevent alt abuse
* prevent one-hit tagging

==================================================================
SECTION 19 — PROTECTED BOSS CHEST LAW
=====================================

Canonical rule:
> *“Personal boss rewards are never dropped directly in combat zones.”*

Boss rewards are delivered through protected reward chests.

Chest properties:
* safe-zone access
* personal ownership
* protected rewards

Open loot does NOT apply to: boss rewards, raid rewards, special event rewards.
Open loot DOES apply to: player corpses, world loot, regular monster corpses.


==================================================================
SECTION 20 — CANONICAL BIBLES INTEGRATION
=========================================

To ensure absolute system alignment, the following canonical Bibles are officially integrated into the system's core design rules:

### 1. Primordial Bible
* Each Primordial represents a unique cosmic force of the universe.
* The **Main Continent** remains a neutral continent with no parent Primordial.

### 2. Elemental Affinity Bible
* **Playable Affinities:** Fire, Ice, Holy, Shadow, and Nature.
* **Non-Playable Affinities:** The **Abyss** (or Void) is NOT a playable affinity.
* **Awakening Requirements:** Permanent Elemental Awakening is unlocked under two simultaneous conditions:
  1. Character Level ≥ 100
  2. Affinity Level = 100
* **Awakening Lock:** The first awakened affinity is permanent and irrevocable. Secondary awakenings are strictly prohibited.

### 3. Experience Curve Bible
* **Character Level Max:** 9999
* **Progression Curve Speed:**
  * **Levels 1–100:** Fast-paced progression (onboarding and base vocational growth).
  * **Levels 100–200:** Medium-paced progression (awakening and specialization).
  * **Levels 400+:** Slow endurance progression (designed for long-term player dedication).

### 4. PvE Combat Bible
* **High TTK (Time-To-Kill):** Combat encounters are slow, tactical, and deliberate.
* **Low Monster Density:** Fights focus on individual high-threat monsters or small cohorts rather than massive, mindless mob swarms.
* **Auto-Attack Speed:** Basic combat attacks trigger automatically according to weapon attack speeds.
* **Low-to-Medium Cooldown Spells:** Active spellcasting rotates around small, active tactical cooldowns.
* **Mana Recovery via Hits:** Spellcasters and warriors recover mana dynamically through successful basic attacks (auto attacks) and physical hits.
* **Active Defense Systems:** Combat demands active mitigation through block, parry, dodge, and dynamic defensive recovery systems.


==================================================================
SECTION 21 — WORLD ACTIVITY MATRIX BIBLE
==================================================================

### 1. System Overview
The world of Light and Shadow MMORPG is structured as an open-ended, non-linear activity ecosystem. Player progression, movement, and choices are entirely self-directed. Player engagement emerges from selectable activity types, each defining its own risk profile, reward scaling, encounter density, economic contribution, and death exposure.

### 2. Core Design Principle
The central paradigm of player agency is enshrined in the following rule:

<text id="activity-freedom-principle">
Players may engage in any available activity at any time, with no enforced progression sequence.
</text>

All player choices flow through a unified risk-reward cycle:
`Risk → Encounter → Outcome → Economic Impact → Progression Shift`

### 3. World Activity Categories
The open sandbox is classified into four canonical activity classes:

#### A. HUNT ACTIVITIES
* **Purpose:** Combat-driven resource gathering, trophy hunting, and leveling.
* **Risk Level:** High / Extreme.
* **Properties:**
  * Controlled, hunt-based spawn encounters override geographical barriers.
  * Low default monster density with high TTK (Time-to-Kill) combat requiring deliberate tactical pacing.
  * Loot rolls and corpse interaction (Inventory Overflow Law) fully enabled.
  * High exposure to PvE death penalties.
* **Examples:** Solo hunts, group hunting parties, elite sub-regions, demon-infested ruins, and dragon territories.

#### B. EXPLORATION ACTIVITIES
* **Purpose:** World discovery, cartography, and secret finding.
* **Risk Level:** Minimal / Low.
* **Properties:**
  * Very low combat density.
  * Indirect, abstract rewards.
  * Minimal death exposure; optimized for active navigation and world traversal.
* **Examples:** Charting unexplored areas, discovering hidden zones, locating historical landmarks, and uncovering regional secrets.

<text id="exploration-reward-neutral">
Exploration rewards are abstract and undefined pending future canonical Crafting Bible definition.
</text>

#### C. QUEST / CONTRACT ACTIVITIES
* **Purpose:** Structured, objective-based narrative and economic progression.
* **Risk Level:** Moderate (dependent on target zone and objective difficulty).
* **Properties:**
  * Clearly defined objectives, tasks, and bounty targets.
  * Direct completion rewards in standard currencies and materials.
  * May dynamically spawn tactical encounters or hunts.
* **Examples:** NPC contracts, bounty tasks, guild trade runs, and regional event contracts.

#### D. BOSS / WORLD EVENT ACTIVITIES
* **Purpose:** Apex group challenges, server-scale coordination, and legendary loot acquisition.
* **Risk Level:** Extreme Potential (variable by encounter).
* **Properties:**
  * Multi-phase boss fight profiles with complex mechanics.
  * Hybrid participation validation (e.g. contribution and activity duration checks).
  * Personal boss rewards drop directly into secure personal chests (Protected Boss Chest Law) inside safe zones.
  * No open corpse loot available for boss-level rewards.
* **Examples:** World raid bosses, server-wide demonic invasions, sovereign dragon manifestations, and dungeon final bosses.

<text id="boss-risk-principle">
Boss encounters represent the highest complexity and reward tier, but do not define guaranteed death probability.
</text>

### 4. Activity Risk Model
All world activities are governed by a qualitative risk tier structure rather than artificial numeric representations:

<text id="risk-tier-system">
Activities are classified by qualitative risk tiers: Minimal, Low, Moderate, High, Extreme.
</text>

The coupled risk-reward matrix is defined by:
* **Death Risk Rating:** Qualitative risk tier dictates the baseline threat.
* **Loot Exposure Level:** Qualitative risk tier governs likelihood of item drop on PvE death.
* **Encounter Density:** Low / Medium / High.
* **TTK Expectation:** Seconds to minutes scale (dictates tempo).
* **Blessing Relevance Factor:** Essential (influences mitigation needs).

This relationship is formalized under the following canonical rule:

<text id="risk-reward-binding">
Higher risk activities must proportionally increase reward potential through loot quality, currency yield, or crafting material density.
</text>

### 5. Economic Integration Rule
Every activity type must funnel back into the game's core economic cycle:
* **Currency Hierarchy:** Bronze → Silver → Gold → Diamond Coins are earned proportionally to task difficulty.
* **Fixed Item Identity:** Items are immutable templates. They cannot be enhanced or mutated.
* **Crafting Dominance:** 60% of the game's economic weight is anchored in crafting, necessitating steady material flow from hunts.
* **Scarcity Model:** Adheres to the "B+" philosophy, ensuring high-quality items remain rare and highly valued assets.

### 6. Death System Integration
Activities that carry a danger rating must interface seamlessly with the canonical death laws:
* Hardcore PvE death penalties (risk of dropping equipped gear and backpack items).
* Open Loot Corpse System (dropped items become lootable by anyone on the map).
* Inventory Overflow Corpse Persistence (unclaimed items remain on the corpse for up to 12 minutes).

This integration is governed by the following canonical rule:

<text id="death-activity-link">
Activities must explicitly define how death alters reward loss, loot exposure, and economic consequence.
</text>

### 7. Blessing System Integration
The blessing system acts as a player-funded buffer for risky activities:
* Exactly 7 Blessings exist in total.
* **Linear Protection:** Scaled directly via `Final Death Loss = Base Death Penalty * (1 - Active Blessings / 7)`.
* **Zero Loss Threshold:** Having all 7 active blessings guarantees 100% protection from item/currency drops.
* **Consumption Law:** Any PvE death immediately purges all blessings to 0.

This is formalized as:

<text id="blessing-risk-modifier">
Blessings function as a global risk mitigation modifier applied across all activity types.
</text>

### 8. Sandbox World Design Philosophy
In alignment with the non-linear design, the game eliminates arbitrary character-level progression gating in favor of:

<text id="sandbox-activity-model">
The world is a sandbox of independent activity nodes governed by risk, reward, and encounter configuration rather than player level pathing.
</text>

This is supported by the following canonical rule:

<text id="region-danger-principle">
Regions are not level-locked. All regions are accessible at any level, but naturally vary in danger, encounter density, and risk.
</text>

Danger is emergent, not restricted. Instead of "level-locked" regions, players navigate the world assessing risk versus reward, choosing whether to enter high-threat zones with or without blessing insurance, and engaging in gameplay loops aligned with their immediate objectives.

==================================================================
CRAFTING & ITEM TIER BIBLE — FINAL CANON
========================================

### 1. Core System Principle
Crafting in the Light and Shadow MMORPG is a foundational pillar designed for deep, deterministic progression.

<text id="crafting-core-principle">
Crafting is a universal system available to all players without profession restrictions, governed by material acquisition and recipe knowledge.
</text>

Every player can craft any item in the game up to Tier 3, provided they acquire the specific deterministic materials and discover the respective recipe template.

### 2. Item Tier Structure (1–5)
All gear, weapons, accessories, and armor in the world are organized into exactly 5 qualitative tiers:
* **Tier 1:** Early game starter equipment
* **Tier 2:** Low-mid game progression equipment
* **Tier 3:** Mid game transition equipment
* **Tier 4:** Late game master equipment
* **Tier 5:** Endgame apex legendary equipment

This structural organization is governed by the following canonical rules:

<text id="tier-definition-rule">
Item tiers define acquisition method, not item quality or stat variance.
</text>

<text id="item-template-rule">
Items are defined as abstract deterministic templates without narrative naming unless explicitly registered in canonical item database.
</text>

### 3. Crafting Eligibility Rule
To preserve high-end item scarcity and a thriving late-game loot hunt, crafting is strictly bounded by item tiers:
* **Tier 1:** Fully Craftable
* **Tier 2:** Fully Craftable
* **Tier 3:** Craftable (Hybrid transition tier)
* **Tier 4:** Non-Craftable (Drop-Only)
* **Tier 5:** Non-Craftable (Drop-Only)

This boundary is formalized under the following canonical rule:

<text id="crafting-tier-limit">
Items of Tier 4 and Tier 5 cannot be crafted under any circumstances.
</text>

### 4. Item Drop Integration
How different item tiers enter the world is strictly governed to maintain high encounter relevance:
* **Tiers 1 & 2:** Primarily crafted by players, though basic monsters may occasionally drop basic items.
* **Tier 3:** Hybrid acquisition. Players may craft them through high-tier materials, or discover them as rare drops from elite monsters.
* **Tiers 4 & 5:** Exclusively obtained as loot drops from high-tier activities, bosses, elite encounters, and sovereign dragon/demon manifestations.

### 5. Recipe Acquisition System
To prevent instantaneous crafting progression, recipes are not automatically known upon starting.

<text id="recipe-unlock-canon">
Recipes are unlocked exclusively through quests and structured narrative progression events.
</text>

No external system or world event may grant recipes.

### 6. Universal Crafting System
The system is built on democratic, open access:

<text id="universal-crafting-rule">
Crafting capability is universal; complexity is determined only by recipe requirements and material availability.
</text>

There are no crafting classes, blacksmithing/alchemy specializations, or restrictive profession levels. Any player who meets the material and recipe requirements can craft any eligible item.

### 7. Material System Integration
Deterministic materials reside at the center of the crafting economy:

<text id="material-source-rule">
Materials are exclusively obtained from combat-based World Activities (Hunts, Bosses, Elite encounters).
</text>

IMPORTANT:
* No gathering systems (mining, woodcutting, herbalism) exist.
* No environmental extraction mechanics are implemented.
* No passive world resource nodes are available.
* All crafting materials originate as combat rewards from eligible encounters.

### 8. Progression Interaction
The game rejects arbitrary enhancement systems:

<text id="no-upgrade-crafting-rule">
Crafting produces only base deterministic items with no enhancement or modification systems.
</text>

All item progression is replacement-based and economy-driven. Items produced are static, immutable templates following the Fixed Item Identity Law.

### 9. Economic Structure Alignment
The economic shift across tiers ensures long-term system stability:
* **Early game (T1-T2):** Completely dominated by player crafting, fostering a highly collaborative and trading-rich early player economy.
* **Mid game (T3):** Transition hybrid economy where craft meets rare drops.
* **Endgame (T4-T5):** Complete shift to drop-only scarcity, mitigating mudflation, protecting core reward significance, and establishing a definitive high-tier loot hierarchy.

### 10. Final Verification Parameters
The following constraints must be met by all operational subsystems:
* Under no circumstances can a Tier 4 or Tier 5 item be produced via crafting.
* Crafting success must remain 100% deterministic, producing exact immutable templates.
* Recipes must map to valid quest or progression milestones.
* No professions or crafting specializations may exist.
* All economics must respect the Bronze, Silver, Gold, Diamond Coin hierarchy.


## Canonical PvP Combat & World Conflict Bible (Final Canon Version)

This section defines the official and unified rules governing PvP Combat and World Conflict in **Light and Shadow**. It functions as a complete, authoritative system extension that maintains full, unbroken parity with the existing PvE Combat, Death Penalty, and Blessing structures.

### 1. System Overview & Unified Mechanics
The PvP system is designed as an elegant, non-fragmented extension of the core game world rather than a separate gameplay mode. All player vs. player combat, deaths, loot drops, and economic rules share direct mechanical parity with player vs. monster systems.

<text id="pvp-pve-parity-rule">
PvP and PvE share identical combat, death, loot, and penalty systems with no mechanical divergence.
</text>

There are no dedicated PvP progression routes, specialized PvP gears, stat modifications (e.g., resilience stats), or separate PvP currencies. Combat performance is determined entirely by identical elemental alignments, item templates, and player positioning.

### 2. Safe Zone Structure
PvP interaction is completely and unconditionally restricted inside designated cities, outposts, and protected hubs. 

<text id="safe-zone-rule">
Cities and designated safe hubs completely disable all forms of PvP interaction.
</text>

Inside these coordinates:
* PvP damage scaling is forced to 0%.
* Aggressive combat initiation is physically blocked.
* Corpse looting of any kind is disabled.
* This ensures that players have absolute safety for trade, recovery, and socializing.

### 3. Open World PvP Rules
Outside designated safe zones, the entire world is an active conflict area where PvP is fully enabled.

<text id="open-world-pvp-rule">
All non-safe zones are PvP-enabled and function under full open-world combat conditions.
</text>

* Player killing is a fully emergent sandbox element of the world.
* There are no arbitrary server-enforced flags, faction locks, or alignment restrictions to restrict combat initiation.
* Combat engagement is completely emergent, dynamic, and player-driven.

### 4. Death System Unification
When a player falls in PvP combat, the exact same death consequence engine triggers as in a PvE death. There is absolutely no mechanical divergence.

<text id="pvp-death-unification">
PvP and PvE deaths are mechanically identical in terms of item loss, corpse generation, and economic consequences.
</text>

Every death results in:
* Generates a lootable player corpse in the open world.
* Subject to standard hardcore item drop rules (equipped and inventory item loss based on active blessings).
* Adheres to the standard Inventory Overflow rule (remaining items stay in corpse storage).
* Generates standard death penalties without artificial PvP-only scaling.

### 5. Blessing System Integration
The 7 Blessings Linear Protection Model applies equally and universally to both PvP and PvE deaths.

<text id="blessing-unified-rule">
Blessings mitigate both PvP and PvE death penalties using a unified linear system and are fully consumed upon death.
</text>

* **0 to 7 Active Blessings:** Each active blessing linearly mitigates 14.285% of item loss penalty, up to 100% full item protection at 7 blessings.
* **Universal Blessing Consumption:** Any player death, regardless of whether it was caused by a monster (PvE) or another player (PvP), instantly consumes ALL active blessings. The player must re-visit Altars to restore protection.

### 6. Player Corpse Loot Rules
Player corpses created through PvP encounters conform entirely to the standard open loot laws.

<text id="corpse-loot-unified-rule">
All player corpses are globally lootable regardless of PvP or PvE origin.
</text>

* Player corpses persist for up to 12 minutes in the open world if they contain unlooted items.
* Standard Open Loot rules apply: the corpse can be looted by the killer, allies, strangers, or opportunistic scavengers who reach it.
* Player corpses are not protected or locked, creating intense field security scenarios around active combat zones.

### 7. Economic Role & Impact of PvP
Rather than acting as an isolated gameplay loop, PvP functions as a primary driver of the global item-driven sandbox economy:

<text id="pvp-economic-role">
PvP is an economic redistribution system integrated into the global loot and death economy.
</text>

* **Wealth Redistribution:** Equips, materials, and coins are transferred directly from defeated players to victorious players or field scavengers, acting as a natural balance of power.
* **Risk & Circulation Amplifier:** Open-world PvP increases item circulation velocity and drives high-end template replacement demand.
* **Blessing Consumption Sink:** Every death completely consumes active blessings, reinforcing continuous coin demand as players pay altar tribute to restore protection.
* **Scarcity Reinforcement:** Increases the scarcity of high-tier templates (especially drop-only Tier 4 and Tier 5) by removing them from circulation during high-risk activities.

### 8. Combat Initiation Rules
Initiating PvP follows clean mechanical triggers:
* **Direct Attack:** Targeting another player and executing a damage-dealing or crowd-control action outside a safe zone.
* **Retaliation:** Defending yourself or allies from an active aggressor, which flags you as a combat participant.
* **Zone-Based Engagement:** Entering hot combat activities (e.g., world bosses, elite hunts) where engagement risk is inherently assumed.
* **Safe Zone Failsafe:** Fleeing into a safe zone completely blocks ongoing PvP damage and closes initiation paths.

### 9. Design Philosophy & Boundaries
To keep **Light and Shadow** a unified, highly polished world, PvP is never allowed to fragment the community or game design:

<text id="pvp-design-boundary">
PvP exists as an extension of PvE systems, not as a parallel system.
</text>

There is no parallel track of PvP content. All content is simply world content, and PvP is the natural high-stakes human friction layer layered on top.

### 10. PvP Validation Parameters
To ensure perfect integrity, the following conditions must be guaranteed across all operational systems:
1. All PvP deaths trigger identical hardcore item loss as PvE.
2. The Blessing protection percentage is identical and linear across both.
3. Death from a player or monster consumes all blessings in exactly the same way.
4. Designated cities and safe hubs completely block PvP damage.
5. No separate PvP gear progression or stats exist.


## Canonical Quest & Contract Bible (Final Canon Version)

This section establishes the authoritative system design and architecture for the Quest and Contract system in **Light and Shadow**. It acts as a primary gameplay framework designed to preserve complete sandbox freedom while providing deep world lore and repeatable loops.

### 1. System Overview & Architecture
The quest and contract systems are distinct, coexisting paths. Together, they populate the world of **Light and Shadow** with optional narrative threads and infinite repeatable gameplay tasks.

<text id="quest-system-core">
The game uses both persistent quest content and repeatable contract content as parallel world activity systems.
</text>

Importantly, quests are never a mandatory choke point for character progression. A player can reach maximum capabilities, master elemental alignments, and acquire high-end gear completely through other world activities.

<text id="activity-freedom-quest-rule">
Quests and contracts are optional world activities and never impose mandatory linear progression paths.
</text>

This supports the ultimate sandbox identity, where player choice determines their daily activity matrix (hunts, exploration, bosses, PvP, or contracts).

### 2. Canonical Quest Classification System
All persistent quest content in the world is classified into one of four rigid canonical quest classes, each with distinct design parameters:

* **Main Quests:** Focused on macro-world progression, major story arcs, continent access (such as accessing the Holy Continent), and unlocking major boss/dungeon hubs.
* **Story Quests:** Dedicated to world lore, the history of the Primordials, ancient elemental conflicts, and faction narratives. These are fully optional and expand narrative depth without restricting system access.
* **Unlock Quests:** Act as gates for specific gameplay tools, such as unlocking specific spells, rare crafting recipes, or specialty hunting fields, without forcing a rigid path.
* **Side Quests:** Abundant local narratives that enrich village depth, support minor NPCs, and reward exploring off-the-beaten-path locations.

<text id="quest-freedom-principle">
Quests are abundant world activities and are not limited to progression gating. Players may pursue quests for progression, lore, rewards, exploration, or personal goals.
</text>

### 3. Contract System (Repeatable Loop)
Contracts are designed to sustain repeatable, long-term, infinite world-interaction loops. Unlike narrative-focused quests, contracts can be executed endlessly.

<text id="contract-repeatable-rule">
Contracts are repeatable tasks that provide long-term gameplay loops independent of story progression.
</text>

The game provides five standard, repeatable contract archetypes:
* **Hunt Contracts:** Simple extermination tasks targeting regional monster populations.
* **Bounty Contracts:** Tracking and defeating a dangerous named elite creature with dynamic spawn markers.
* **Escort Contracts:** Defending trading caravans or NPC diplomats traveling across non-safe zone routes where emergent PvP risks are high.
* **Delivery Contracts:** Safely transporting fragile trade goods or scrolls between remote outposts.
* **Elite Kill Contracts:** Challenging localized mini-bosses or specialized elite commanders.

### 4. Quest Reward Matrix
Completing quests and contracts distributes diverse, deterministic rewards into the economy:

<text id="quest-reward-rule">
Quest rewards may include progression, economy, knowledge, and unlock-based rewards.
</text>

<text id="quest-reward-template-rule">
Quest rewards use deterministic reward templates selected from fixed canonical reward pools.
</text>

Authorized rewards are strictly defined as:
* **Experience (XP):** Character progression increments.
* **Currency:** Payed in the standard copper, bronze, silver, gold, and diamond coin hierarchy.
* **Crafting Materials:** Essential tier templates (T1, T2, T3) needed for blacksmithing.
* **Deterministic Items & Recipes:** Fixed template gear or blueprints with fixed, non-mutable values. No random attributes or mutable quality rolls.
* **Spell Unlocks:** Forbidden or rare discipline spells.
* **Faction Reputation:** Measuring alignment status with NPC factions.

### 5. Experience Integration & Source Priority
To preserve the classic, slow, and grind-oriented identity of **Light and Shadow**, experience distribution is carefully prioritized:

<text id="xp-source-priority-rule">
Hunts remain the primary source of character experience while quests provide secondary progression rewards.
</text>

Quests and contracts provide solid auxiliary boosts, but active, tactical field hunts remain the dominant method of leveling up elemental alignments.

### 6. NPC Factions
Factions are structured world organizations, serving as the central hubs for contract boards and specialized skill learning.

<text id="npc-faction-rule">
NPC factions function as progression, learning, and contract hubs rather than player social organizations.
</text>

<text id="faction-extensibility-rule">
NPC factions are not limited to a fixed number and may expand dynamically as the world evolves through future content, continents, and narrative systems.
</text>

The listed factions function as canonical starter factions and do not represent a hard numerical limit on the world's faction architecture. These starter examples include:
1. **Hunter Lodges:** Trackers and beast slayers offering Hunt/Bounty contracts and beast-slaying gear templates.
2. **Mage Orders:** Scholar circles providing elemental spell training, magical research quests, and arcane lore.
3. **Religious Orders:** Wardens of holy temple altars, providing blessing services and holy contracts.
4. **Mercenary Guilds:** Soldier boards offering Bounty, Delivery, and Escort contracts for coins.
5. **Shadow Cults:** Dark, hidden operations offering forbidden spells, dark lore, and high-risk contracts in open-world conflict zones.

### 7. Hybrid Spell Learning System
Unlocking and advancing character spells uses a hybrid model:

<text id="spell-learning-hybrid-rule">
Spells are learned through both faction training and quest-based unlocks depending on spell rarity and narrative significance.
</text>

* **Faction-Learned Spells:** Common discipline and combat utility spells purchased directly from faction trainers.
* **Quest-Learned Spells:** Unique, ancient, or highly specialized elemental spells unlocked exclusively as rewards for completing difficult Main, Story, or Unlock quests.

### 8. Failure and Persistence Policy
To prevent permanent progression blockers and protect player efforts:

<text id="quest-persistence-rule">
Quests never fail permanently and remain recoverable regardless of interruption or abandonment.
</text>

* **Non-destructive Failure:** If a player dies or abandons an objective, they can instantly restart or resume it later.
* **NPC Immortality:** Quest-givers and target NPCs are physically immortal and immune to player-killing, guaranteeing that story nodes remain fully operational.

### 9. Sandbox Design Boundaries
Above all, the Quest & Contract systems must remain subservient to sandbox freedom:

<text id="quest-design-boundary">
Quest systems must support sandbox freedom and may never replace open-world activity choice as the core gameplay philosophy.
</text>

The game is not a linear theme-park ride. Quests provide flavor and goals, but the freedom to define your own path in the open world remains absolute.

### 10. Quest System Validation checklist
All implementers must ensure:
1. Four canonical quest categories (Main, Story, Unlock, Side) are defined.
2. Contracts are repeatable and separate from narrative quests.
3. Hunts are confirmed as the primary XP source, with quests serving as auxiliary rewards.
4. Factions exist as npc hubs rather than player guilds.
5. Spell learning is hybrid (trainer-based + quest-unlocked).
6. Quests can never fail permanently or brick character progression.


## Guild & Social Bible (Final Canon Version)

This section establishes the authoritative design and mechanics for the **Guild and Social System** in **Light and Shadow**. Guilds are intentionally constructed as small-scale social coordination groups, designed to support friendship, group activities, and sandbox cooperation without introducing rigid political, territorial, or economic mechanics that can lead to large-group domination.

### 1. Core Guild Philosophy
In **Light and Shadow**, guilds are lightweight vehicles for social gathering and coordination. They are completely dissociated from geopolitical world status, territorial warfare, or monopolistic trade coalitions.

<text id="guild-core-principle">
Guilds function as small-scale social coordination groups and are not territorial, political, or economic control entities.
</text>

Players may form and join guilds to easily organize hunts, coordinate contracts, explore high-danger dungeons, conquer world bosses, or coordinate open-world activities. However, guilds never dictate character attributes, level scaling, or exclusive region locks.

### 2. Strict Membership Scale Limits
To foster clear social identity, prevent massive zerg dominance, and ensure tight-knit community cohesion:

<text id="guild-size-limit">
Guilds are limited to a maximum of 30 members.
</text>

This hard numerical cap is absolute. It prevents the emergence of giant mega-guilds that dilute individual reputation and force uncoordinated gameplay.

### 3. Simplified Rank Structure
The internal management of guilds is standardized and direct. Customized hierarchy levels are omitted to keep structure flat and low-overhead:

<text id="guild-rank-rule">
Guilds use a fixed three-rank hierarchy: Leader, Officer, Member.
</text>

This three-tier authority model handles basic invites, kicks, and guild house customization privileges, avoiding unnecessary corporate simulation.

### 4. Absence of Systemic Alliances
Emergent diplomacy is preserved through the absence of code-enforced alliances:

<text id="guild-alliance-rule">
There are no formal alliance systems between guilds.
</text>

If separate guilds wish to cooperate, they must do so informally through verbal agreements and player-driven social trusts, preventing the establishment of unyielding multi-guild cartels.

### 5. Private Economic Integrity (No Shared Storage)
To completely prevent the centralized consolidation of wealth, guild-tax systems, and administrative item management overhead:

<text id="guild-storage-rule">
Guilds do not have shared banks or collective storage systems.
</text>

Every item and coin belongs exclusively to the player who acquired or crafted it. There are no collective guild vaults. Wealth remains distributed individually, supporting the sandbox barter economy.

### 6. Sandbox Risk and Trust (Guild PvP)
Guild membership does not create system-enforced physical safety or friendly-fire immunity outside safe zones:

<text id="guild-pvp-rule">
Guild membership does not prevent PvP interactions outside safe zones.
</text>

This preserves high-stakes emergent drama. Players must actively build and manage trust. Betrayals can happen in high-danger territory, making social cohesion a real gameplay skill.

### 7. Zero-Cost Guild Creation
Creating a guild represents a social contract rather than an economic privilege:

<text id="guild-creation-rule">
Guild creation has no currency cost and is freely available to eligible players.
</text>

This removes financial barriers, enabling any group of friends or solitary trackers to instantly establish their own social tag.

### 8. Guild Houses (The Gold Sinks)
The only structural asset a guild can acquire is a Guild House. In strict alignment with the standard currency sink requirements:

<text id="guild-house-acquisition-rule">
Guild houses are acquired exclusively through gold payments.
</text>

Guild Houses serve purely as cozy social and non-combat utility centers:

<text id="guild-house-rule">
Guild Houses function as social and utility hubs providing services such as crafting stations, NPC interactions, and quest/contract boards.
</text>

Within a Guild House, members can gather, place trophy achievements, utilize private crafting stations, interact with vendor NPCs, and access unified contract boards.

### 9. Sandbox Design Boundaries
The system is intentionally constrained to prevent power consolidation and meta-progression bottlenecks:

<text id="guild-design-boundary">
Guild systems are designed as lightweight social structures and must not extend into territorial control or global economic dominance.
</text>

By design, a solo player or a small circle of three players suffers no gameplay penalties or attribute deficits compared to a member of a maximum-sized guild.

### 10. Guild System Validation Checklist
Implementers must ensure the following boundaries are upheld:
1. Max guild size is exactly 30 members.
2. Formal alliance systems are completely absent.
3. Collective guild banks/storage do not exist.
4. PvP remains fully enabled between guild members in active open-world conflict zones.
5. Creating a guild has zero currency cost.
6. Guild houses are bought with Gold only.
7. Guild houses provide utility & social amenities (crafting stations, trophy rooms, faction boards).
8. Guild systems have zero mechanical influence on territorial capture or global commerce.


## Trade & Market Bible (Final Canon Version)

This section establishes the authoritative design and mechanics for the **Trade and Market System** in **Light and Shadow**. The game features a fully player-driven, friction-balanced, hybrid trade economy engineered to promote circulation, safe assets storage, and gold-sink mechanisms to sustain economic integrity.

### 1. Core Economic Philosophy
The economy of **Light and Shadow** relies on high-trust secure transactions, emergent low-trust interactions, and global asynchronous trading hubs:

<text id="trade-core-principle">
The economy operates through hybrid trade channels combining direct trade, manual exchange, and a global asynchronous market.
</text>

This multi-faceted structure ensures that trading adapts to any playstyle while upholding sandbox economic freedom.

### 2. Secure Direct Trade
For high-trust player-to-player exchanges, characters can initiate direct trading via an interactive interface:

<text id="secure-trade-rule">
Secure trade uses a protected confirmation window requiring approval from both participants.
</text>

Any manual change to the items or currency offered immediately resets all prior approvals to prevent bait-and-switch exploits.

### 3. Emergent Manual Handoff
For raw sandbox encounters, players can exchange items in a completely unmoderated fashion:

<text id="manual-trade-rule">
The game supports unprotected manual item exchange that carries theft and betrayal risk.
</text>

Dropping items onto the floor or swapping them in unsecured zones carries inherent risk. It represents a valid, emergent form of transaction with zero system protections.

### 4. Globally Connected Markets
Every populated city features a trade market situated near the Depot (DP). These markets are globally unified:

<text id="global-market-rule">
All city markets are globally connected and share the same listing pool.
</text>

A buyer in a remote safe-zone town sees the exact same listings as a buyer in the bustling Capital, maximizing liquidity and item circulation.

### 5. Listing Inventory Deposit
When placing an item on the market, it leaves the player's ownership immediately:

<text id="market-deposit-rule">
Listed items are removed from player inventory and held by the market until sold or cancelled.
</text>

This prevents players from listing an item while continuing to use it in combat, keeping the listing database synchronized with real available inventory.

### 6. Fixed Price Instant Purchases
The market relies entirely on direct, transparent transactions:

<text id="market-pricing-rule">
Market listings use fixed-price instant purchases and do not support bidding or auction systems.
</text>

Bidding wars, dynamic extensions, and complex auction rules are omitted to maintain immediate and fluid trading.

### 7. Non-Refundable Listing Fees (The Gold Sinks)
To combat database clutter and listing spam, and to act as a permanent deflationary sink:

<text id="market-listing-fee-rule">
An upfront non-refundable listing fee of 2% of the listing price is charged when an item is listed on the market.
</text>

This gold-based listing fee is paid upfront, represents a vital currency sink, and is non-refundable.

### 8. Indefinite Listing Durations
To preserve player listing intent and support long-term asynchronous trading without high administrative maintenance:

<text id="market-duration-rule">
Market listings remain active indefinitely until sold or manually cancelled.
</text>

Listings do not expire, allowing rare craft templates to find buyers naturally over time.

### 9. Cancellation Policy
If a player changes their mind, they can cancel a listing, but they pay the price:

<text id="market-cancellation-rule">
Cancelled listings return the item to the player's depot while market fees remain permanently consumed.
</text>

The item is sent back to the local depot, but the upfront fee remains lost, deterring constant pricing manipulation.

### 10. Universal Item Tradability (No Soulbound)
In strict alignment with the **Fixed Item Identity Law**, items are not artificially bound to characters:

<text id="market-tradability-rule">
All items in the game are fully tradable across all trade channels.
</text>

There are no bind-on-pickup (BOP), soulbound, or bind-on-equip (BOE) states. Any weapon, armor, legendary recipe, or boss material remains a tradable asset.

### 11. Automated Bank Transfers
When a sale is successfully processed, the seller receives their coins safely:

<text id="market-payment-delivery-rule">
Gold from successful market sales is transferred directly into the character bank.
</text>

This prevents seller characters from being killed on the way to the market with a pockets full of high-value gold.

### 12. Carried Currency Risk
When exploring the high-risk zones, physical currency is in danger:

<text id="carried-currency-risk-rule">
Only currency physically carried by the character is subject to death loss.
</text>

If a player carries gold coins in their active backpack, a portion is dropped or lost on death, following the linear blessing protection formula.

### 13. Secure Bank Protection
To enable safe wealth accumulation, cities provide risk-free depository boxes:

<text id="bank-security-rule">
Currency stored in the character bank is fully protected from all death penalties.
</text>

Banked currency is entirely immune to death penalties, incentivizing players to return to safe zones to deposit their earnings.

### 14. Sandbox Economic Design Boundaries
To preserve a player-first economy where interaction remains central:

<text id="trade-design-boundary">
Trade systems must preserve player-driven economy while maintaining meaningful social interaction and economic risk.
</text>

System-controlled merchants do not set prices, and artificial item scarcity is determined by hunt drop chances and crafting recipes, not synthetic market gates.


## Race & Character Creation Bible (Final Canon Version)

This section establishes the authoritative design, limitations, and onboarding framework for all playable races and newly created characters within **Light and Shadow**. Character creation in this world balances roleplay diversity, deep visual aesthetics, and absolute competitive parity.

### 1. System Overview
Playable races in **Light and Shadow** are designed purely to define visual identity, lore, cultural backgrounds, and immersive aesthetic variety:

<text id="race-core-principle">
Playable races exist for identity, aesthetics, and lore and do not create mechanical power differences.
</text>

No race holds mechanical or scaling superiorities over another, ensuring a level playing field.

### 2. Available Playable Races
The roster of selectable races comprises exactly five distinct options, each deeply embedded in the historical and environmental lore of the regions:

<text id="playable-races-rule">
The game supports five playable races: Forest Elf, Human, Dwarf, Ice Elf, and Green Orc.
</text>

These five groups represent the civilized, cooperative factions who have established peace inside the core neutral regions.

### 3. Absolute Statistical Neutrality
To prevent meta-gaming and ensure that race selection remains purely a personal aesthetic choice:

<text id="race-neutrality-rule">
Playable races provide no mechanical combat, progression, or statistical advantages.
</text>

All characters, regardless of race, share the exact same base HP, Mana, movement speed, damage scaling, and progression attributes. There are zero racial traits, racial passives, or elemental resistances that differentiate them numerically.

### 4. Shared Onboarding Safe Zone
To consolidate early population and build a cooperative multiplayer atmosphere right from the start:

<text id="starting-zone-rule">
All playable races begin in the same starting region.
</text>

No race is sequestered in an isolated island or distant continent; everyone shares the same early-game journey.

### 5. Universal Starting City
Every newly created character, regardless of race or chosen class, starts their adventure inside the same neutral citadel:

<text id="ironhold-starting-city-rule">
All newly created characters begin in the neutral starting city of Ironhold Bastion.
</text>

This bastion contains essential secure systems such as Depot storage, banking terminals, early quest boards, safe zone boundaries, and global trade market terminals.

### 6. Absolute Class Freedom
Racial boundaries do not dictate mechanical vocations. Every race is fully capable of studying any path:

<text id="class-race-freedom-rule">
All playable races may choose any available class or vocation.
</text>

Whether a stout Dwarf desires to be a Sorcerer, or a heavy Green Orc is drawn to the path of a Cleric, the game imposes no class restrictions.

### 7. Equal Racial Reputation
Factions and cities across the continent look past a character's ancestry:

<text id="racial-reputation-rule">
Race does not influence faction reputation, NPC behavior, or city access.
</text>

All cities and shops remain fully open and equally responsive to all characters, regardless of race.

### 8. Racial Permanence
Racial choices are absolute. In alignment with competitive stability:

<text id="racial-permanence-rule">
Chosen race is permanent and does not evolve into higher racial forms.
</text>

There are no dynamic racial evolutions, hidden subrace promotions, or mid-game race-change cash-shops that compromise character consistency.

### 9. Gameplay Hitbox Standardization
To guarantee absolute fairness in both PvE positioning and high-stakes PvP combat:

<text id="racial-hitbox-rule">
All playable races use identical gameplay hitboxes regardless of visual body proportions.
</text>

Whether a character is an imposing Green Orc or a compact Dwarf, their physical targeting cylinder, collision volume, and pathfinding footprints are identical in the physics engine.

### 10. Dual-Gender Visual Options
Character customization supports binary gender options affecting representation only:

<text id="gender-selection-rule">
Character creation supports male and female gender selection.
</text>

Gender selection does not affect base power, inventory space, or combat mechanics.

### 11. Core Design Boundaries
To prevent the creeping of meta advantages and focus on world immersion:

<text id="race-design-boundary">
Race systems must enhance identity and immersion without affecting gameplay balance.
</text>

Racial differences manifest solely through custom-tailored cosmetics, culturally themed starter clothing, voiceovers, and unique animations, never through numbers.


## Class & Vocation Bible (Final Canon Version)

This section establishes the authoritative design, delayed specialization, mechanical progression, and equipment restriction parameters of the class and vocation systems of **Light and Shadow**.

### 1. Core Philosophy
To allow deep experimentation before choosing a specialized focus:

<text id="class-core-principle">
Characters begin classless and later specialize into permanent combat classes that define exclusive skills, spells, and progression paths.
</text>

All new players learn the foundational mechanics of movement, positioning, and generic resource handling prior to specializing.

### 2. The Novice Phase (Level 1–9)
Upon stepping into the world, characters do not hold any traditional class designations:

<text id="novice-phase-rule">
All characters remain classless novices between levels 1 and 9.
</text>

This Novice state represents a tabula rasa, preventing players from being locked into archetypes before understanding basic mechanics.

### 3. Novice Weapon Freedom
To encourage thorough testing of all early combat styles:

<text id="novice-weapon-rule">
Classless characters between levels 1 and 9 may equip and use any starter weapon available in Ironhold Bastion.
</text>

Novices can freely swap between daggers, swords, bows, staves, or axes to see which mechanics suit them, before specializing.

### 4. Specialization Selection
Upon reaching level 10, characters must choose their vocation:

<text id="class-selection-rule">
At level 10, every character must permanently select one of the five playable classes.
</text>

The five selectable classes are: **Knight**, **Mage**, **Archer**, **Assassin**, and **Cleric**.

### 5. Selection Permanence
Commitment to a vocation is absolute and final:

<text id="class-permanence-rule">
Class selection at level 10 is permanent and cannot be changed afterward.
</text>

The game does not offer class resets or vocation swaps, enforcing deep character identity and class-specific reputation.

### 6. Gear-Driven Combat Roles
While classes grant specific toolkits, the items you wear define how you use them:

<text id="gear-driven-class-rule">
Class identity defines access to abilities and progression paths, but combat role is primarily determined by equipped items rather than fixed class statistics alone.
</text>

A Knight's build may shift completely from a high-armor tank to an offensive damage dealer depending on their choice of armor and weaponry.

### 7. Absolute Skill & Spell Exclusivity
To make each class feel mathematically unique and specialized:

<text id="class-skill-exclusivity-rule">
Each class possesses exclusive skills and spells that cannot be learned or used by other classes.
</text>

There is no cross-class skill pooling or multi-class hybridizing. Mage spells belong exclusively to Mages, and healing prayers belong exclusively to Clerics.

### 8. Item-Based Class Restrictions
Instead of broad weapon-category hard-locks, compatibility is governed by item-specific metadata:

<text id="item-class-requirement-rule">
Items may contain explicit class requirements defined in their template metadata. Only compatible classes may equip or use those items.
</text>

Certain items like a "Staff of Frost" require the "Mage" class tag in character sheets, while a "Common Dagger" remains completely unrestricted.

### 9. Stat-Growth & Attribute Differentiation
When the character selection is confirmed at level 10, foundational stats are permanently branched:

<text id="class-base-attribute-rule">
Each class possesses unique base attributes assigned upon class selection at level 10, defining its initial combat strengths and weaknesses.
</text>

Each class starts with distinct and optimized base statistics (HP, Mana, Defense, Speed, Magical scaling) that define their core combat identity.

### 10. Racial Neutrality Preservation
Class selection does not dilute the identity-only design of races:

<text id="class-race-neutrality-rule">
Class selection must never override race neutrality or create race-based class restrictions.
</text>

Every race can select every class. An Orc Mage has the exact same stat profile and spell availability as an Elf Mage, ensuring absolute competitive balance.


## Spell & Skill Bible (Final Canon Version)

This section establishes the authoritative design, execution dynamics, learning mechanics, and scaling systems for spells and active combat skills in **Light and Shadow**.

### 1. Core Philosophy
Skills and spells are the primary mechanical differentiators of combat roles, providing active tactical flavor:

<text id="spell-core-principle">
Skills and spells define combat execution and class identity while remaining fully integrated with the gear-driven combat model.
</text>

Active skills complement the tier-based items wielder equips, blending action and progression smoothly.

### 2. Absolute Class Exclusivity
To preserve class uniqueness and encourage team synergy:

<text id="class-skill-exclusivity-rule">
Each class possesses exclusive skills and spells that cannot be learned or used by other classes.
</text>

Cross-class learning is forbidden. A Knight can never study Frost Magic, and an Assassin can never heal using Divine prayers.

### 3. Active Combat Mana Regeneration
To reward aggressive, tactical positioning and prevent passive stalls:

<text id="mana-regeneration-rule">
Mana is primarily regenerated through active combat interactions such as successful attacks, spell hits, and combat-triggered effects.
</text>

Characters cannot rely on out-of-combat passive mana ticks; they must engage in combat to sustain their casting loops.

### 4. Automated Basic Attacks
To optimize focus on skill timing and position:

<text id="auto-attack-rule">
Basic attacks are automatic while valid hostile targets remain inside combat range and line of sight.
</text>

The physical combat engine automatically swings or shoots when targets are close, letting players concentrate on situational active abilities.

### 5. Fast-Paced Low Cooldowns
To ensure engaging and active moment-to-moment combat:

<text id="spell-cooldown-rule">
All spells and active skills possess low cooldowns designed to support fast tactical decision-making without excessive downtime.
</text>

Skills have short, rapid cooling sequences rather than multi-minute lockouts, minimizing tedious waiting times.

### 6. Instant Casting System
To guarantee fluid movement and responsive gameplay:

<text id="instant-cast-rule">
All spells and active skills are instant cast with no casting time or channeling delay.
</text>

No spells require standing still to fill cast bars, maintaining the quick pacing of the isometric environment.

### 7. Skill Progression by Repeated Usage
Progression relies entirely on mastery through repetition:

<text id="skill-progression-rule">
Skills and spell proficiencies increase through repeated use rather than manual point allocation.
</text>

Swinging swords naturally increases Sword proficiency, while healing teammates naturally increases Holy/Healing mastery.

### 8. Unlimited Hotbar Skill Access
To reward comprehensive learning and keep combat setups versatile:

<text id="skill-access-rule">
Characters are not restricted by hotbar slot limits and may access all learned skills and spells.
</text>

The game provides full accessibility to a player's complete arsenal, structured through clean menus or customized keys.

### 9. Dual Acquisition Methods
Spells are obtained through in-world progression and tutoring:

<text id="spell-learning-rule">
Spells are learned through NPC trainers or quest completion depending on spell rarity and importance.
</text>

Common skills are taught by Bastion trainers, while legendary spells require finishing hazardous regional quests.

### 10. Multi-Layered Skill Scaling Model
Final combat output is highly dynamic and multi-layered, avoiding single-attribute dominance:

<text id="skill-scaling-rule">
Final spell and skill output is determined by base skill power, class scaling, skill proficiency, equipment scaling, and elemental modifiers.
</text>

Calculated as: `FinalOutput = (BaseSkillPower + SkillProficiency * 2.5) * ClassCoefficient * EquipmentMultiplier * ElementalAffinitiesModifier`. This ensures stats, player experience, gear, and elements work in perfect harmony.


## Content Generation & World Expansion Framework (Canonical Version)

This section establishes the official guidelines, systems, and constraints for expanding world content (creatures, environmental zones, instances, global bosses, and dynamic world activities) in **Light and Shadow**.

### 1. Activity-Driven Sandbox Philosophy
World progression is fluid and activity-driven, prioritizing user choice over artificial gating:

<text id="content-core-principle">
World content is defined by systemic risk, encounter design, and activity context rather than player level or linear progression gating.
</text>

Players explore according to risk appetite rather than numeric status locks.

### 2. Multi-Dimensional Encounter Generation
Rather than restrictive level-bracket locks on areas:

<text id="encounter-generation-rule">
Encounters are generated based on biome logic and activity type rather than fixed regional or level-based constraints.
</text>

Every environment dynamically loads creature configurations, group spacing, and aggressive tactics suitable to current conditions.

### 3. Preserving the Canonical Bestiary
To maintain the core identity of the creature families:

<text id="bestiary-content-rule">
Creatures are assigned behavior profiles and encounter contexts but do not alter or redefine canonical bestiary family structures.
</text>

Content expansions append modifiers, behaviors, or scaling levels to the core families (Demon, Dragon, Beast, Undead) rather than polluting the taxonomy with arbitrary new types.

### 4. Spatiotemporal Dynamic Spawn System
Spawning is fluid, realistic, and contextual:

<text id="spawn-system-rule">
Creature spawns are dynamic and driven by activity context, biome conditions, and world state rather than fixed location rules.
</text>

There are no static camp-and-kill spawners. Entities migrate or manifest depending on active regional hunt parameters.

### 5. Multi-Phase Global World Bosses & Events
Bosses are massive regional milestones that cross static boundaries:

<text id="boss-event-rule">
Bosses and world events are global systemic encounters that can manifest in any biome based on world state conditions.
</text>

When a rift erupts or a world event initiates, specific high-difficulty bosses alter the environment and challenge local players.

### 6. Biome Conditions & Affinity Influences
Environmental factors shape the combat theater:

<text id="biome-rule">
Biomes influence encounter behavior and rewards but do not restrict creature or boss types.
</text>

A fiery wasteland increases fire elemental strengths and biases drops toward fire protection gear, but does not forbid ice creatures or general loot.

### 7. Risk-Tier Scaling Model
Difficulty is categorized in a standardized risk-tier hierarchy:

<text id="content-scaling-rule">
Content difficulty is determined by systemic risk tiers and encounter complexity rather than player level or static region classification.
</text>

Tiers range from Minimal Risk (safe outer limits) to Extreme Risk (untamed rifts).

### 8. Structured Dungeon & Instance Framework
Dungeons offer optional challenges tailored to organized squads:

<text id="dungeon-framework-rule">
Dungeons are optional structured encounters layered on top of the open world system and do not restrict sandbox freedom.
</text>

Dungeon chains scale with team composition, ensuring accessible challenges for solo, duo, or full team parties without enforcing linear requirements.


## World Scale Bible (Canonical Integration)

This section details the canonical constraints, physical/spatial parameters, and settlement dimensions that define the grand theater of **Light and Shadow**.

### 1. World Unit Scale
* **Base Metric:** 1 tile = 1 meter (fully canonical).

### 2. Standard Continent Scale
All major continents share identical physical dimensions to establish a consistent, epic spatial foundation:

<text id="continent-base-scale-rule">
All standard major continents measure 12 kilometers by 12 kilometers, equivalent to 12,000 by 12,000 traversable tiles.
</text>

* **Applies To:** Main Continent, Fire Continent, Ice Continent, Holy Continent, Shadow Continent, and Nature Continent.
* **Derived Metrics:** 
  * Width/Height: 12,000 tiles (12,000 meters / 12 km)
  * Total Area: 144 km²

### 3. Abyssia Scale Exception
Abyssia, as the high-danger endgame island of floating anomalies and void rifts, breaks standard geographic limits:

<text id="abyssia-scale-rule">
Abyssia is the only continent with expanded physical dimensions, measuring 16 kilometers by 16 kilometers, equivalent to 16,000 by 16,000 traversable tiles.
</text>

* **Derived Metrics:** 
  * Width/Height: 16,000 tiles (16,000 meters / 16 km)
  * Total Area: 256 km²

### 4. Continent Accessibility
While continents feature vast mountain ranges, water bodies, and mystical blockades, they are highly traversable to encourage endless exploration:

<text id="continent-accessibility-rule">
Approximately 90% of each continent's physical surface is traversable and explorable by players. Only 10% may be inaccessible due to natural barriers, environmental hazards, protected regions, or world-boundary restrictions.
</text>

* **Inaccessible Barriers:** Steep cliffs, active lava lakes, glacial cliffs, bottomless abyss cracks, sealed sanctuaries, or void boundary walls.
* **Derived Scale Ranges:**
  * **Standard Continents:** 129.6 km² of explorable land, 14.4 km² of natural barriers.
  * **Abyssia:** 230.4 km² of explorable land, 25.6 km² of natural barriers.

### 5. Continent Traversal Time
Due to the vast scale of the terrain, crossing continents on foot is a significant undertaking:

<text id="continent-traversal-rule">
Traversing a standard continent from edge to edge via ground movement should require approximately 45–60 minutes of continuous travel.
</text>

* **Abyssia Traversal:** Crossing Abyssia from edge to edge on foot requires approximately 70–90 minutes of continuous ground movement.

### 6. Mount Mobility Framework
Ground mounts exist to assist adventurers but are purely dedicated to travel and travel-efficiency:

<text id="mount-mobility-rule">
Ground mounts provide moderate movement speed bonuses exclusively for travel. Mounts never grant combat or progression advantages.
</text>

* **Multiplier:**
  * **Basic Mounts:** Provide a 125% base movement speed bonus.
  * **Epic Mounts:** Provide a 140% base movement speed bonus.
* **Limitations:** Mounts do not grant attributes, skills, resistances, spell/combat stats, or progression advantages of any kind.

### 7. Logical Coordinate Systems vs. Physical Dimensions
The coordinate grid serves operational and networking functions:

<text id="logical-coordinate-rule">
Region coordinates define logical placement anchors for streaming, spawning, and teleportation systems and must not be interpreted as literal physical continent boundaries.
</text>

* Real-time spatial tracking relies on logical anchor positions mapped out in configuration libraries.

### 8. World Layout and Geographic Orientation
The climate and geography of the continents obey macro-environmental logic:

<text id="world-layout-rule">
Continents are distributed freely across world space without radial or symmetrical constraints. Ice Continent is predominantly northern and Fire Continent predominantly southern.
</text>

* This ensures that while spatial layout is organic and sandbox-oriented, the environmental themes follow intuitive global temperature vectors.


## Settlement Scale Bible

Every city, outpost, and fortress is built according to a rigid scale classification to ensure immersive and readable proportions across the sandbox world.

### 1. Small City / Small Settlement
These are small outposts, frontier defense forts, and starter encampments designed for swift traversals and specialized functions.

<text id="small-settlement-scale-rule">
Small settlements measure approximately 400 to 700 tiles in width and height.
</text>

* **Derived Scale:**
  * Traversal Time: 2–5 minutes of continuous travel from edge to edge.
  * **Applies To:** Ironhold Bastion, Last Bastion.

### 2. Medium City
The standard regional hubs, specialized trade towns, and coastal ports that drive the continental barter economy.

<text id="medium-settlement-scale-rule">
Medium cities measure approximately 800 to 1500 tiles in width and height.
</text>

* **Derived Scale:**
  * Traversal Time: 5–10 minutes of continuous travel from edge to edge.
  * **Applies To:** Thornwall, Blackwater Bay, Crimson Hollow, Molten Anvil, Frosthaven, Lunareth, Sunwall, Grimharbor, Kar’goth, Vel’Sharum, Sylvaris, Oakenspire, Grunhold.

### 3. Large City
Monumental fortress complexes and ancestral strongholds carved directly into natural wonders.

<text id="large-settlement-scale-rule">
Large cities measure approximately 1600 to 2500 tiles in width and height.
</text>

* **Derived Scale:**
  * Traversal Time: 10–18 minutes of continuous travel from edge to edge.
  * **Applies To:** Stone Tirith, Ymirr’s Hidden Cavern.

### 4. Capital City
The epicentres of political power, arcane research, and supreme commerce, standing as the ultimate urban achievements of each race.

<text id="capital-settlement-scale-rule">
Capital cities measure approximately 2500 to 4000 tiles in width and height.
</text>

* **Derived Scale:**
  * Traversal Time: 15–30 minutes of continuous travel from edge to edge.
  * **Applies To:** Ravenshire, Pyra Magnus, Elarisheim, Luminaar, Noctharyn, Necrathis, Elarin.


## Itemization Bible

The itemization architecture of **Light and Shadow** defines a structured, fair, and immersive equipment economy, prioritizing strategic skill choice and voxel-precision combat while providing deeply rewarding, non-inflated progression loops.

### 1. Core Principle

<text id="item-core-principle">
Items enhance combat power and build identity but must not override skill execution or class identity.
</text>

* Combat is a balanced combination of class abilities, tactical choices, player movement, and gear parameters. Equipment amplifies these variables but never completely bypasses player competence.

### 2. Item Rarity Tiers

<text id="item-tier-rule">
The game uses exactly 5 item tiers.
</text>

* **Tiers:** Tier 1, Tier 2, Tier 3, Tier 4, and Tier 5.
* **No Alternate Labels:** To avoid typical MMO vertical inflation and rarity slop, there are no artificial rarity labels such as "Common", "Rare", "Epic", or "Legendary". The Tier number encapsulates both rarity and base item power brackets.

### 3. Tier Power Scaling

<text id="tier-power-rule">
Higher tiers generally provide stronger stats and more valuable affixes.
</text>

* Power increases incrementally with each Tier (Tier 1 < Tier 2 < Tier 3 < Tier 4 < Tier 5) to support healthy progression without causing runaway power creep or exponential mathematical bloat.

### 4. Tier 5 Artifacts

<text id="tier-5-rule">
Tier 5 items are nearly legendary artifacts.
</text>

* Tier 5 items represent the peak of ancient power and forgotten relics. They are extremely rare, highly prestigious, and act as highly stable stores of value in the player trade economy. They never drop as common loot.

### 5. Crafting Restraints

<text id="crafting-tier-rule">
Crafting is restricted to Tier 1, Tier 2, and Tier 3. Tier 4 and Tier 5 items are not craftable.
</text>

* Highly specialized armorers and blacksmiths can fabricate reliable gear up to Tier 3. Tiers 4 and 5 represent items lost to antiquity or forged by elemental forces, obtainable only through active world exploration and epic achievements.

### 6. Upgrade Restrictions

<text id="item-upgrade-rule">
Items CANNOT be upgraded or evolved.
</text>

* To respect the investment of players and protect sandbox item scarcity, all items are final upon creation/drop. There are no enhancement stars, refinement systems (+1 to +7), awakening paths, or systems that evolve items into higher tiers.

### 7. Horizontal Itemization & Diversity

<text id="horizontal-itemization-rule">
The system must support large horizontal item variety.
</text>

* Items within the same tier feature highly distinct stat distributions rather than simple linear vertical upgrades. One Tier 4 sword might maximize physical damage, another might emphasize crit chance, while a third focuses on elemental fire damage to synergize with Mage or Knight builds.

### 8. Weapon Categories

<text id="weapon-category-rule">
Canonical weapon categories are restricted to: Sword, Axe, Bow, Dagger, Staff, and Shield.
</text>

* These 6 specific weapon families map natively to class restrictions, custom weapon masteries, and combat output formulas.

### 9. Equipment Slots

<text id="equipment-slot-rule">
Canonical equipment slots are restricted to: Weapon, Shield / Off-hand, Helmet, Armor, Legs, Boots, Ring 1, Ring 2, Amulet, and Cape.
</text>

* Every character has access to these 10 discrete item slots, offering a rich canvas for multi-slot builds.

### 10. Class Requirements

<text id="class-item-requirement-rule">
Items may have explicit class restrictions.
</text>

* Certain items require specific vocational mastery (e.g., Staves for Mages, heavy tower shields or massive axes for Knights). Race-based equipment restrictions are strictly forbidden to ensure cross-racial build equality.

### 11. Elemental Affinities

<text id="elemental-affinity-item-rule">
Items may contain elemental offensive or defensive modifiers.
</text>

* Equipment is natively woven into the game's elemental affinity system. Modifiers include offensive bonuses (Fire, Ice, Holy, Shadow, Nature Attack) and defensive resistances (Fire, Ice, Holy, Shadow, Nature Resist).

### 12. World Acquisition (Drops)

<text id="item-drop-rule">
Items may enter the world through monster drops, boss drops, world events, contracts, and future dungeons.
</text>

* Standard enemies drop lower-tier gear, whereas high-danger world bosses, elite challenges, and catastrophic events serve as the exclusive gateway for Tier 5 artifacts.

### 13. Market Integration

<text id="market-compatibility-rule">
All items must integrate with Trade & Market Bible.
</text>

* All weapons, armor, and accessories are fully tradable, market-compatible, and securely exchangeable between players to support a thriving, organic sandbox economy.

### 14. Death Penalty Safety

<text id="death-penalty-item-rule">
Only carried items are exposed to death-loss risk.
</text>

* Items held in player vault banks are 100% safe from death penalties, maintaining a clear distinction between active risk-carrying inventories and long-term secure deposits.

### 15. Combat Output & Damage Calculation

To govern combat formulas consistently, physical classes (especially Knights, Archers, and Assassins) rely heavily on weapon base statistics:

$$\text{Final Combat Output} = \text{Skill Base Power} \times \text{Class Scaling} \times \text{Weapon Base Power} \times \text{Equipment Affixes} \times \text{Elemental Affinity Modifiers}$$

This ensures that finding or crafting an exceptionally balanced higher-tier weapon directly translates to combat dominance, while still remaining tethered to skill and class parameters.

---

# ITEMIZATION EXPANSION PATCH (FINAL CANON VERSION)

## 1. ARMOR ARCHETYPES BIBLE

### System Philosophy
Armor is divided into archetypes that define defensive identity and playstyle specialization without breaking class balance.

### Canonical Rules

<text id="armor-archetype-rule">
There are exactly three canonical armor archetypes: Heavy Armor, Light Armor, and Cloth Armor.
</text>

<text id="heavy-armor-rule">
Heavy Armor provides the highest physical defense, low dodge, moderate to low magic resistance, and prioritizes survivability in frontline combat.
</text>

<text id="light-armor-rule">
Light Armor provides medium physical defense, high dodge, mobility-oriented bonuses, and offensive utility.
</text>

<text id="cloth-armor-rule">
Cloth Armor provides low physical defense, high magic resistance, high mana synergy, and strong spell-oriented scaling.
</text>

### Class Alignment

* Knight → primarily Heavy Armor
* Archer → primarily Light Armor
* Assassin → primarily Light Armor
* Mage → primarily Cloth Armor
* Cleric → primarily Cloth Armor

*Note: Armor archetypes influence item identity and affix weighting but DO NOT create hard race restrictions.*

---

## 2. WEAPON ARCHETYPES BIBLE

### System Philosophy
Weapons define combat identity, scaling patterns, attack profiles, and build diversity. The weapon system is split into two independent layers: Weapon Type and Damage Archetype.

### Canonical Rule

<text id="weapon-archetype-rule">
There are exactly ten canonical weapon types in Light and Shadow.
</text>

### Layer 1 — Weapon Type

* **MELEE**
  * **Sword**: Balanced physical weapon with versatile offensive profile.
  * **Axe**: Slow heavy weapon with high burst damage.
  * **Mace**: Blunt weapon specialized in stagger and anti-heavy-armor combat.
  * **Spear**: Extended melee reach and zoning capability.
  * **Dagger**: Very high attack speed and crit-oriented burst.
* **RANGED**
  * **Bow**: Long-range sustained physical DPS.
  * **Crossbow**: Slower than bow but higher burst and armor penetration.
* **MAGIC**
  * **Staff**: Primary spell-power weapon with strong AoE scaling.
  * **Wand**: Fast-casting magical weapon with mana efficiency.
  * **Tome / Relic**: Support-oriented magical focus for healing and utility scaling.

### Layer 2 — Damage Archetype

<text id="damage-archetype-rule">
Damage archetypes define how damage is calculated independently of weapon shape.
</text>

The canonical damage archetypes are:
* **Slashing**: Emphasizes clean cutting, high bleeding, and persistent wound damage. (Sword, Axe)
* **Piercing**: High critical hit rate and deep structural penetration. (Spear, Dagger)
* **Bludgeoning**: High stagger, guard break, and effective penetration of Heavy Armor. (Mace)
* **Magical**: Channels pure elemental or spiritual damage, scaling with intelligence and spell-power. (Staff, Wand, Tome / Relic)
* **Ranged Physical**: High velocity ranged physical damage, scaling with distance and dexterity. (Bow, Crossbow)

Every equippable weapon item maps to exactly one **Weapon Type** and exactly one **Damage Archetype**.

---

## 3. DUAL WIELD BIBLE

### System Philosophy
Dual wield is an exclusive class mechanic that increases offensive potential in exchange for survivability or utility.

### Canonical Rules

<text id="dual-wield-rule">
Dual wield is only available to explicitly authorized classes and weapon combinations.
</text>

### Allowed Dual Wield Configurations

* **Knight**:
  * Sword + Shield
  * OR
  * Dual Sword
* **Assassin**:
  * Dagger + Offhand
  * OR
  * Dual Dagger

*No other class may dual wield.*

### Damage Formula

<text id="dual-wield-final-rule">
Offhand weapon contributes exactly 75% of its total damage value.
Main hand contributes 100% of its damage value.
</text>

* **Formula**:
  $$\text{Effective Weapon Power} = (\text{MainHandPower} \times 1.0) + (\text{OffHandPower} \times 0.75)$$

### Tradeoff Rules

* **Knight Dual Sword**:
  * Higher burst, Higher DPS
  * Loses shield defense, Lower survivability
* **Assassin Dual Dagger**:
  * Faster burst combos, Higher crit frequency
  * Lower utility, Extremely fragile

*Note: Dual wield is NOT a free upgrade.*

---

## 4. AFFIX POOL BIBLE

### System Philosophy
Items derive build diversity from modular stat affixes while preserving base item identity.

### Canonical Rules

<text id="affix-pool-rule">
All equippable combat items may roll attributes from the global canonical affix pool.
</text>

### Affix Pool Categorization

* **OFFENSIVE AFFIXES**: ATK, Crit Chance, Crit Damage, Armor Penetration, Attack Speed, Skill Power
* **DEFENSIVE AFFIXES**: Armor, Magic Resist, Dodge, Block Chance, Max HP
* **RESOURCE AFFIXES**: Max Mana, Mana Regeneration, HP Regeneration, Cooldown Reduction
* **ELEMENTAL OFFENSE**: Fire Attack, Ice Attack, Holy Attack, Shadow Attack, Nature Attack
* **ELEMENTAL DEFENSE**: Fire Resist, Ice Resist, Holy Resist, Shadow Resist, Nature Resist

---

## 5. ITEM IDENTITY RULES

<text id="item-identity-rule">
Affixes may augment an item but may never invalidate its base identity.
</text>

* **Examples**:
  * A sword remains a physical weapon.
  * A staff remains a magic weapon.
  * Cloth cannot become heavy armor.

---

## 6. TIER / AFFIX SCALING RULE

<text id="tier-affix-scaling-rule">
Item tier determines affix count, affix magnitude, and elemental roll probability.
</text>

### Tier Rules

* **Tier 1**: 1 to 2 minor affixes
* **Tier 2**: 2 to 3 affixes
* **Tier 3**: 3 to 4 moderate affixes
* **Tier 4**: 4 to 5 powerful affixes
* **Tier 5**: 5 to 6 major affixes (Near-artifact rarity)

---

## 7. COMBAT FORMULA & DEFENSE BIBLE (FINAL CANON VERSION)

### System Philosophy
Light and Shadow is a strictly gear-driven combat MMORPG. Characters do not have standard RPG attributes like Strength, Dexterity, Intelligence, or Vitality. All offensive output and defensive performance are governed purely by equipment, skill powers, elemental affinities, resistances, and active combat modifiers. This design ensures that high Time-to-Kill (TTK) and deep tactical play remain paramount.

### Canonical Rules

<text id="combat-core-principle">
Combat output in Light and Shadow is fully determined by equipped gear, skill proficiency, skill power, elemental scaling, and combat modifiers rather than traditional attribute systems. Characters do NOT gain damage or scaling from Strength, Dexterity, Intelligence, Wisdom, Agility, or any other hidden character attribute.
</text>

<text id="high-ttk-combat-rule">
Combat pacing is intentionally slow and tactical, with survivability heavily influenced by gear and mitigation systems while preventing infinite defensive scaling.
</text>

<text id="final-damage-formula">
Final Damage = (Weapon Damage + Skill Damage) × Elemental Modifier × Critical Modifier (if critical hit).
Where Weapon Damage is the fixed damage value of the equipped weapon, Skill Damage is the fixed base power from the spell or skill cast, Elemental Modifier is derived from the net elemental interaction, and Critical Modifier is applied exclusively on critical hits.
</text>

<text id="armor-diminishing-returns-rule">
Physical armor mitigation follows a diminishing returns curve to preserve survivability while preventing invulnerability.
Damage Reduction = Armor / (Armor + K)
Where Armor is total physical armor from gear and K is the global balancing constant (K = 250). Mitigation increases quickly at low values, slows at high values, and never reaches 100%. Tanks remain durable without becoming immortal. Legacy references to K=500 are removed.
</text>

<text id="horizontal-armor-budget-rule">
Armor values are not determined by rigid tier ranges. Each item possesses unique stat distribution based on power budget allocation. Items may trade armor for critical chance, elemental resistances, affinity bonuses, utility affixes, or niche PvP/PvE specialization.
</text>

<text id="tier5-bis-armor-anchor-rule">
A full Best-in-Slot Tier 5 tank-oriented armor set is balanced around approximately 500 total armor from the entire equipped defensive set (e.g., Helmet ≈ 70, Chest ≈ 160, Legs ≈ 120, Boots ≈ 50, Shield ≈ 100).
</text>

<text id="mitigation-benchmark-rule">
A fully optimized tank build should usually remain around 60–70% physical mitigation, preserving high survivability without reaching invulnerability.
</text>

<text id="elemental-resistance-rule">
Elemental resistances from gear (Fire Resist, Ice Resist, Holy Resist, Shadow Resist, Nature Resist) reduce incoming elemental damage after physical mitigation calculations, stacking from equipped gear to support precise combat calculations.
</text>

<text id="affinity-defense-rule">
When a character reaches maximum awakened elemental affinity (Level 100) for an element, the character gains an additional +4% passive defensive resistance against that matching element. This bonus is passive and stacks with item elemental resistance.
</text>

<text id="critical-hit-rule">
The base critical hit chance for all classes is 5%. Normal classes have a critical damage multiplier of 1.5x damage, while the Assassin class has an enhanced critical damage multiplier of 2.2x damage. Critical damage applies after elemental scaling calculations: Critical Damage = Final Damage × Critical Multiplier.
</text>

<text id="dual-wield-offhand-rule">
Offhand weapon contributes exactly 75% of its total damage value, while the main hand contributes 100%. Allowed classes are Knight (dual swords) and Assassin (dual daggers).
</text>

### Class Survivability Targets
These targets define expected mitigation levels for optimized mid-to-late-game builds across the primary vocations:
* **Knight**: 60–70% mitigation (Highly durable front-line tank)
* **Cleric**: 25–45% mitigation (Tactical dynamic combat support)
* **Archer**: 20–35% mitigation (Evasive ranged combatant)
* **Assassin**: 15–30% mitigation (High-risk critical strikes)
* **Mage**: 10–25% mitigation (Glass-cannon elemental burst caster)

### Formula Implementations & Mathematical Walkthroughs

1. **Offensive Output**:
   $$\text{Final Damage} = (\text{Weapon Damage} + \text{Skill Damage}) \times \text{Elemental Modifier} \times \text{Critical Modifier}$$
   * Note: Elemental Modifier is computed as $1.0 + (\text{AttackerAffinity} - \text{DefenderResistance}) \times 0.01$.

2. **Physical Armor Damage Reduction**:
   $$\text{Damage Reduction} = \frac{\text{Armor}}{\text{Armor} + 250}$$
   * Canonical Benchmarks:
     * $\text{Armor} = 100 \implies \text{Damage Reduction} = 100 / 350 \approx 28.57\%$
     * $\text{Armor} = 250 \implies \text{Damage Reduction} = 250 / 500 = 50.0\%$
     * $\text{Armor} = 500 \implies \text{Damage Reduction} = 500 / 750 \approx 66.67\%$
     * $\text{Armor} = 700 \implies \text{Damage Reduction} = 700 / 950 \approx 73.68\%$

3. **Net Damage Taken**:
   $$\text{Final Damage Taken} = \text{Final Damage} \times (1 - \text{Damage Reduction}) \times (1 - \text{Total Elemental Resistance})$$
   * Where $\text{Total Elemental Resistance}$ is the sum of gear-provided resistances and the $+4\%$ affinity defense bonus (if matching affinity is at Level 100).

## 8. PROGRESSION BIBLE (FINAL CANON VERSION)

### System Philosophy
Character progression in Light and Shadow is entirely skill-driven, gear-influenced, and elemental-scaled, completely bypassing traditional primary attribute models. The attributes Strength, Dexterity, Intelligence, Vitality, Agility, and Luck do not exist in player progression or combat equations.

### Canonical Rules

<text id="attribute-less-progression-rule">
Light and Shadow does not use traditional RPG primary attributes. Character progression is entirely skill-driven.
</text>

<text id="skill-based-power-rule">
Character power is determined by combat proficiencies, equipment quality, and elemental scaling.
</text>

<text id="hybrid-skill-gain-rule">
Skill progression occurs through both real skill usage and contextual mastery XP.
</text>

<text id="anti-macro-skill-rule">
Repeated low-risk repetitive actions provide heavily reduced skill gains.
</text>

<text id="exponential-skill-curve-rule">
Each subsequent skill level requires progressively more effort than the previous level.
</text>

<text id="no-hard-skill-cap-rule">
Combat proficiencies have no hard maximum value and can increase indefinitely.
</text>

<text id="controlled-combat-scaling-rule">
Displayed skill values progress infinitely, but effective combat contribution scales down through prestige bands to prevent runaway power creep.
</text>

<text id="combat-offense-formula-rule">
Offensive output is derived exclusively from weapon base damage, effective skill contribution, elemental modifiers, and critical multipliers.
</text>

### Canonical Skills
The primary character combat proficiencies are:
1. **Sword Fighting**: Governs combat effectiveness when wielding blades.
2. **Axe Fighting**: Governs combat effectiveness when wielding battleaxes.
3. **Club Fighting**: Governs combat effectiveness when wielding maces and clubs.
4. **Dagger Fighting**: Governs combat effectiveness when wielding precision daggers.
5. **Distance Fighting**: Governs combat effectiveness when wielding bows and throwing weapons.
6. **Magic Level**: Governs spell power and restorative magic scaling.
7. **Shielding**: Governs physical block chance and shield performance.

### Exponential Progression Curve
Leveling up skills requires exponential effort relative to the current level:
* **Level 1–30**: Fast progression
* **Level 31–60**: Moderate progression
* **Level 61–90**: Slow progression
* **Level 91–150**: Very slow progression
* **Level 151+**: Extremely slow progression

XP Required for Next Level formula:
$$\text{XP For Next Level}(L) = \lfloor 100 \times 1.08^L \rfloor$$

### Infinite Progression with Controlled Combat Scaling
Displayed skill values grow infinitely, but effective combat contribution is subject to prestige bands to prevent vertical stat inflation:
* **Band 1 (Level 1–150)**: Multiplier is **100%** (1.00)
* **Band 2 (Level 151–250)**: Multiplier is **75%** (0.75)
* **Band 3 (Level 251+)**: Multiplier is **50%** (0.50)

#### Effective Combat Skill Formula
$$\text{Effective Skill}(S) = \begin{cases} 
S & \text{if } S \le 150 \\
150 + (S - 150) \times 0.75 & \text{if } 150 < S \le 250 \\
225 + (S - 250) \times 0.50 & \text{if } S > 250
\end{cases}$$

#### Mathematical Example
A displayed Sword Fighting of **300** results in:
* **Band 1**: First 150 levels provide 150 effective contribution.
* **Band 2**: Next 100 levels (151–250) provide $100 \times 0.75 = 75$ effective contribution.
* **Band 3**: Final 50 levels (251–300) provide $50 \times 0.50 = 25$ effective contribution.
* **Total Effective Combat Skill** = $150 + 75 + 25 = 250.0$ effective levels.

### 8.1 SHIELDING & HYBRID COMBAT XP CORRECTIONS

The following canonical corrections and updates define the unified, production-ready progression systems for *Light and Shadow*.

#### Canonical Rule Registrations

<text id="shielding-skill-rule">
Shielding is a full combat proficiency and progresses through successful defensive combat participation.
</text>

<text id="hybrid-combat-xp-rule">
Combat skill progression uses a three-layer XP model composed of action XP, damage contribution XP, and combat outcome bonus XP.
</text>

<text id="action-xp-rule">
Every valid combat action generates a minimum guaranteed XP contribution.
</text>

<text id="damage-contribution-xp-rule">
Damage contribution XP primarily scales with liquid damage rather than raw damage.
</text>

<text id="combat-outcome-bonus-rule">
Significant combat outcomes provide bonus mastery XP.
</text>

<text id="anti-macro-hard-penalty-rule">
Repetitive low-risk actions receive at least 90 percent XP reduction.
</text>

#### Shielding Skill Specification
Shielding is an official combat proficiency alongside the offensive weapon and magic proficiencies:
* **Sword Fighting**
* **Axe Fighting**
* **Club Fighting**
* **Dagger Fighting**
* **Distance Fighting**
* **Magic Level**
* **Shielding**

Shielding uses the same exponential XP curve and infinite progression formula subject to the prestige bands:
* **Band 1 (Level 1–150)**: 100% effective shielding scaling
* **Band 2 (Level 151–250)**: 75% effective shielding scaling
* **Band 3 (Level 251+)**: 50% effective shielding scaling

##### Shielding XP Sources:
1. **Successful Shield Blocks**: Direct mitigation of an oncoming physical or elemental attack.
2. **Defensive Combat Participation**: Actively standing in the frontline during battles.
3. **Tanking Aggro during PvE**: Generating threat and absorbing attacks from monsters.
4. **Defensive PvP Interactions**: Mitigating active player strikes and absorbing hostile spells.

#### Hybrid Combat XP Model (Three-Layer Generation)
To ensure reliable, exploit-free progression, combat XP is generated from three simultaneous sources:

1. **Layer 1 — Action XP (Guaranteed Baseline)**:
   * Every valid action (sword swing, dagger strike, bow shot, spell cast, or successful block) generates a baseline baseline progression XP.
   * *Formula*: $\text{Action XP} = \text{Base Action Coefficient} \times \text{Skill Weight}$

2. **Layer 2 — Damage Contribution XP (Performance Scaling)**:
   * Scales directly with the real combat impact, prioritizing **Liquid Damage** over Raw Damage.
   * *Raw Damage*: Outgoing pre-mitigation damage from offensive equations.
   * *Mitigated Damage*: Damage prevented by target armor, shield, and defense skills.
   * *Liquid Damage*: Real damage dealt (HP removed from target) after all mitigations.
   * *Formula*: $\text{Damage XP} = \text{Liquid Damage} \times 0.25 + \text{Mitigated Damage} \times 0.10$

3. **Layer 3 — Combat Outcome Bonus XP (Milestone Rewards)**:
   * Grants substantial bonuses upon completing specific combat events.
   * *PvE Milestones*: Elite mob kills, miniboss/boss defeats, dungeon clears.
   * *PvP Milestones*: Kill participation, assists, and capturing strategic objectives.

#### Anti-Macro Hard Penalty Reinforcement
Low-risk, repetitive, or fully automated actions (e.g., training dummy macro loops, low-threat stationary farming) trigger a severe progressive penalty:
* **Abuse Trigger**: Continued actions against targets below threat threshold or in a static position.
* **XP Penalty**: **90% to 95%** reduction in all progression XP.
* *Formula*: $\text{Effective XP} = \text{Total XP} \times (1 - \text{Penalty Rate}) \quad \text{where } \text{Penalty Rate} \ge 0.90$


### 9.0 VOCATION / CLASS BIBLE (LOCKED CANON CONSOLIDATION)

This section provides the absolute, definitive, and locked specification of player classes/vocations, the fully attribute-less mechanics, the single-resource mana system, and the dynamic vector-based performance metrics of *Light and Shadow*.

#### Canonical Rules Registry

```yaml id="voc-lock-001"
five-vocation-rule:
  text: "The game contains exactly five canonical vocations: Knight, Assassin, Archer, Mage, Cleric."
```

```yaml id="attr-lock-001"
attribute-less-rule:
  text: "The game does not use traditional RPG attributes. Progression is skill-based and item-driven only."
```

```yaml id="mana-lock-001"
mana-only-rule:
  text: "All playable vocations use Mana as their sole combat resource."
```

```yaml id="role-lock-001"
emergent-role-rule:
  text: "Combat roles are emergent and derived from equipment and skill configuration, not predefined class roles."
```

```yaml id="vector-lock-001"
role-vector-rule:
  text: "Combat role output must be displayed as numeric performance vectors instead of categorical labels."
```

```yaml id="offhand-lock-001"
offhand-rule:
  text: "Each vocation has strict allowed off-hand equipment defining build identity."
```

```yaml id="base-stat-lock"
base-stats-rule:
  text: "Classes define only base survivability and mana parameters; no scaling attributes exist."
```

#### Absolute Attribute-less System Specification

All traditional RPG primary attributes (such as **Strength**, **Dexterity**, **Intelligence**, **Vitality**, **Agility**, and **Luck**) are completely forbidden and non-existent.

Player power, mitigation, and mechanics are determined exclusively via:
1. **Allowed Base Stats** (Defined purely by vocation):
   * **Base HP**: Baseline survivability capacity.
   * **Base Mana**: Maximum energy available for spellcasting/skill executions.
   * **Mana Regen**: Passive mana points recovered per second.
   * **Base Armor**: Defensive rating reducing oncoming physical damage.
   * **Base Magic Resistance**: Defensive rating reducing oncoming magical damage.
   * **Base Move Speed**: Base travel and positioning rate.
2. **Skill Levels**: Direct progression indices matching used weapon classes and magic domains.
3. **Itemization & Affix Buffs**: Modular increments modifying only the allowed base stats above.

#### Unified Vocation System and Off-hand Equipment Rules

| Vocation | Weapon Restrictions | Strict Allowed Off-Hand Items | Combat Role Emergence Performance Vectors |
| :--- | :--- | :--- | :--- |
| **Knight** | Swords, Axes, Clubs, Daggers | Shield, Sword, Empty | Derived as high Survivability + Sustain, or moderate Survivability + high Damage Output. |
| **Assassin** | Daggers, Swords | Dagger, Shield, Empty | Emerges as maximum physical Damage Output + high Move Speed, or high Survivability in defensive PvP setups. |
| **Archer** | Bows, Crossbows, Daggers | Quiver ONLY | Emerges as high physical/elemental Damage Output + high Utility. Shields are strictly locked and cannot be equipped. |
| **Mage** | Spell Staffs, Daggers | Shield, Spellbook | Emerges as maximum magical Damage Output + high Sustain, or high Survivability + moderate Damage when shield-equipped. |
| **Cleric** | Maces/Clubs, Daggers | Shield, Sacred Scepter | Emerges as maximum Sustain + high Utility, or high Survivability + moderate Sustain when shield-equipped. |

#### Combat Role Output Vector Model

*Light and Shadow* does not assign categorical roles (Tank, DPS, Healer). Player efficiency is represented exclusively as an emergent numeric performance vector mapping tendencies across four axes:
1. **Survivability** (0–100): Physical and magical resilience.
2. **Damage Output** (0–100): Offensive DPS rate (physical and magical).
3. **Utility** (0–100): Tactical crowd control, speed boosts, and positioning advantages.
4. **Sustain** (0–100): Continuous healing potency and resource replenishment efficiency.

#### Monorecurso Mana Policy

All playable vocations depend strictly on **Mana** as their single and active combat resource. There are no secondary gauges, such as Stamina, Rage, Energy, Combo Points, or Fury. Every active combat spell, maneuver, or tactical evasion drains Mana directly, making active resource management universal.


### 10.0 SPELL & SKILL BIBLE (LOCKED COMBAT ENGINE)

This section provides the absolute and definitive canonical specification of player skills, spell combat mechanics, and formula equations for *Light and Shadow*.

#### Canonical Rules Registry

```yaml id="high-ttk-spell-combat-rule"
high-ttk-spell-combat-rule:
  text: "Combat is designed around high survivability, sustained pressure, tactical positioning, and fast reactive spell execution."
```

```yaml id="skill-cooldown-rule"
skill-cooldown-rule:
  text: "Skills are divided into three cooldown categories: Basic (0–4s), Strong (8–15s), Ultimate (15–20s)."
```

```yaml id="global-cooldown-rule"
global-cooldown-rule:
  text: "All active skills trigger a global cooldown of 0.35 seconds to prevent instant burst stacking and macro abuse."
```

```yaml id="instant-cast-rule"
instant-cast-rule:
  text: "All spells and active skills are instant cast. Combat pacing relies on cooldowns, positioning, and mana flow instead of cast-time delays."
```

```yaml id="damage-source-rule"
damage-source-rule:
  text: "Weapon-centric classes primarily deal damage through weapon scaling, while spell-centric classes primarily deal damage through spell scaling."
```

```yaml id="spell-scaling-rule"
spell-scaling-rule:
  text: "Spell power scales from weapon magical base, magic level, and elemental affinity."
```

```yaml id="mana-flow-rule"
mana-flow-rule:
  text: "Mana regenerates passively over time and through successful combat hits, functioning as a pacing regulator rather than a hard combat limiter."
```

```yaml id="healing-scaling-rule"
healing-scaling-rule:
  text: "Healing scales from weapon magical base, magic level, and holy amplification."
```

```yaml id="spell-critical-rule"
spell-critical-rule:
  text: "Spells can critically strike using the global critical hit system."
```

```yaml id="limited-cc-rule"
limited-cc-rule:
  text: "Crowd control exists in limited form and must never dominate combat pacing."
```

```yaml id="aoe-softcap-rule"
aoe-softcap-rule:
  text: "Area damage uses progressive soft target caps to prevent large-scale scaling abuse."
```

```yaml id="buff-category-rule"
buff-category-rule:
  text: "Buffs stack by category exclusivity rather than unrestricted stacking."
```

```yaml id="party-protection-rule"
party-protection-rule:
  text: "Party members are protected from allied AoE damage. Non-party targets remain vulnerable."
```

```yaml id="assassin-critical-override-rule"
assassin-critical-override-rule:
  text: "The Assassin critical hit multiplier override (2.2x) applies to every valid outgoing damage source, including physical, magical, skill-based, and elemental damage."
```

```yaml id="skill-taxonomy-rule"
skill-taxonomy-rule:
  text: "Every spell and skill belongs to one canonical category: Damage, Heal, Buff, Debuff, Mobility, or Utility."
```


#### Complete Spell & Skill Specifications

##### 1. Cooldown Tiers & GCD
Spells are structurally designed to flow quickly without locking players in casting animations. 
* **Basic Skills (0–4s CD)**: Low-cost sustain fillers.
* **Strong Skills (8–15s CD)**: Tactical offensive, defensive or movement skills.
* **Ultimate Skills (15–20s CD)**: Major class-identity effects.
* **Global Cooldown (GCD = 0.35s)**: Every active ability sets off a global cooldown of exactly 0.35 seconds. This guarantees high APM combat while preventing instant macro execution.

##### 2. Damage Scaling Mechanics
Traditional scaling statistics (Str, Dex, Int) are banned. Spell power is computed with a clean, 3-variable formula:
$$\text{SpellPower} = \text{WeaponMagicalBase} \times \text{MagicLevel} \times \left(1 + \frac{\text{ElementalAffinityBonus}}{100}\right)$$

##### 3. Healing & High TTK
Healing is highly potent because combat features a high Time-To-Kill (TTK). Healing scales using exactly three variables: weapon magical base, magic level, and holy amplification. Heal formula:
$$\text{Healing Power} = \text{WeaponMagicalBase} \times \text{MagicLevel} \times \text{HolyAmplification}$$
Additionally, healing in PvP works at 100% full efficiency with no debuffs, keeping healing highly tactical in team battles.

##### 4. AoE Soft Target Caps
To mitigate train-pull scaling or massive zerg wipes in Guild wars, AoE spells utilize progressive target-scaling multipliers:
* **1–5 Targets**: 100% damage output
* **6–10 Targets**: 80% damage output
* **11–20 Targets**: 60% damage output
* **21+ Targets**: 40% damage output

##### 5. Party Protection
Allied area of effect damage has a dynamic verification layer. Targets belonging to the caster's party are entirely immune to their AoE hits. In contrast, non-party targets remain fully vulnerable.

##### 6. Exclusive Buff Stacking
To control power creep, players can only benefit from one active buff per categories:
* **Speed Category**: Only one active speed multiplier.
* **Defense Category**: Only one active defensive rating multiplier.
* **Offensive Aura Category**: Only one active aura boost.
Adding a second buff in the same category automatically overwrites or overrides the previous one.

##### 7. Skill Taxonomy
All spells and skills belong to one canonical category to define combat identity:
* **Damage Spell**: Direct offensive damage skills. (e.g. Fireball, Ice Lance, Shadow Burst)
* **Heal Spell**: Direct or over-time healing abilities. (e.g. Heal, Greater Heal, Holy Restoration)
* **Buff Spell**: Positive temporary enhancements. (e.g. Blessing, Holy Shield, Haste)
* **Debuff Spell**: Negative status effects. (e.g. Curse, Slow, Weakness)
* **Mobility Skill**: Movement or repositioning skills. (e.g. Dash, Blink, Leap)
* **Utility Skill**: Non-damage tactical skills. (e.g. Cleanse, Reveal, Mana Transfer)


### 11.0 MONSTER & CREATURE BIBLE (FINAL CANON VERSION)

This section contains the official, absolute, and canonical specification for the Monster, Creature, and World Boss systems of *Light and Shadow*.

#### Canonical Rules Registry

```yaml id="monster-ecosystem-rule"
text: "Monsters are dynamically distributed across the world without fixed tier classification. Difficulty emerges from biome, encounter type, and modifier systems."
```

```yaml id="monster-ai-rule"
text: "Monster AI follows a hybrid model where normal creatures use simple aggro logic while elite and boss entities use tactical decision-making systems."
```

```yaml id="elemental-interaction-rule"
text: "Elemental interactions provide moderate damage modifiers and tactical advantages but do not determine combat outcomes alone."
```

```yaml id="xp-performance-rule"
text: "Experience points are awarded based on player performance contribution during combat encounters rather than fixed monster values."
```

```yaml id="loot-economy-rule"
text: "Loot distribution is structured to support high-TTK gameplay, ensuring sustainable progression without economic inflation."
```

```yaml id="respawn-rule"
text: "Monster respawn is hybrid, combining static timers with dynamic player-driven adjustments to prevent exploitation."
```

```yaml id="boss-system-rule"
text: "Boss encounters exist in two forms: open-world contested bosses and instanced progression bosses."
```


#### Complete Monster System Specifications

##### 1. Core Monster Design Philosophy
In *Light and Shadow*, monsters are part of a living world. There are **no rigid tiers** (e.g., T1–T5 classification) that artificially limit mobs or areas. Instead, a creature's difficulty, stat weight, and behavior are completely emergent, driven by:
* **Biome/Region**: Determines elemental properties, base modifiers, and baseline threat.
* **Encounter Type**: Solo roaming, pack guardians, patrol squads, or event spawns.
* **Elite Modifiers**: Dynamically applied prefixes (e.g., *Stonewarded*, *Flamebound*, *Mana-Gorged*) that scale stats and grant extra abilities.
* **Boss Flags**: World Boss or Instanced Boss states that trigger ultimate mechanics.

##### 2. Hybrid AI System
Creatures in the world operate on a tiered hybrid cognitive engine:
* **Normal Monsters**: Simple, lightweight aggro loops. They chase the highest threat target, attack within range, and hard-reset/leash back to their spawn anchor if pulled too far, healing to full health.
* **Elite & Boss Entities**: Run full **Tactical AI Decisions**. They can:
  - Reposition or backpedal when preparing heavy area attacks.
  - Prioritize softer targets (healers/ranged damage-dealers) if threat levels are comparable.
  - Cycle powerful abilities, deploy shield blocks, or cast interrupts/stuns when players attempt high-cost skills.

##### 3. Elemental Interaction System (Soft Advantage)
Elemental interactions are designed to be tactical and impactful without creating binary, one-shot scenarios.
* Monsters possess moderate resistances and soft elemental vulnerabilities (e.g., +15% damage taken from Fire, -10% from Ice).
* This ensures players are rewarded for swapping skills/weapons to match the target, but combat remains skill-based and balanced, preventing cheese.

##### 4. Performance-Based XP System
Experience point distribution is fully decentralized from fixed, per-kill tables. Players earn XP dynamically based on individual/party combat metrics:
$$\text{XP Earned} = \text{EncounterBaseXP} \times \left( w_1 \cdot \text{DamageContribution\%} + w_2 \cdot \text{HealingContribution\%} + w_3 \cdot \text{ParticipationTime\%} + w_4 \cdot \text{SurvivalModifier} \right)$$
This encourages balanced party roles, rewards active gameplay, and ensures healers/tanks receive equal compensation compared to pure damage classes.

##### 5. Controlled High-TTK Loot Economy
The loot engine supports the high-TTK combat pacing of *Light and Shadow*.
* **Generosity vs. Inflation**: Progression drops are frequent enough to support active gearing, but high-end currencies and legendary items are strictly gated.
* **Drop Influences**:
  - **Biome**: Affects crafting mats, regional currency, and elemental shards.
  - **Monster Type**: Affects armor weight types, weapon frames, and skill grimoires.
  - **Rarity Rolling**: Normal, Magic, Rare, Epic, and Legendary rolls scale the item's baseline item level, random prefixes, and socket allocation.

##### 6. Hybrid Respawn System
To maintain immersion and prevent infinite bot farming in static coordinates:
* Base respawn timers govern each area.
* Timers scale dynamically: when player density is high or a specific camp is heavily hunted, respawn rates adapt (with diminishing yield or elite spawn risk multipliers) to prevent trivialized farming.

##### 7. World Bosses vs. Instanced Bosses
The high-end boss architecture divides encounters into two distinct, balanced avenues:
* **World Bosses**: Spawn in public, contested zones. They encourage massive PvPvE scenarios where rival guilds battle for tap rights while simultaneously dealing with complex, high-health raid mechanics.
* **Instanced Bosses**: Closed, progression-focused encounters. Intended for parties to master deep mechanical triggers, phase shifts, and tactical co-ordination without external disruption.

---

## 🗺️ World Foundation Bible (Pre-World Generation Architecture)

This section contains the authoritative design and integration matrices for the pre-world generation layer of **Light and Shadow**. It binds the Combat, Spell & Skill, Monster, Itemization, and Progression bibles into a unified, non-linear ecosystem.

### 1. Core World Design Philosophy

The world is designed as a living, organic sandbox. Progression is open and fluid, driven by spatial and biological dynamics rather than rigid artificial thresholds.

```yaml id="world-foundation-rule"
text: "The world is a dynamic ecosystem where gameplay systems emerge from biome, encounter density, and systemic interactions rather than static map design."
```

* **Open World**: Traversable without loading screens or gates.
* **Non-linear Progression**: Freedom to choose pathways; level caps are not tied to zones.
* **Biome-driven Difficulty**: Local hazards, density, and monster groupings create natural risk.
* **Systemically Interconnected**: Biome traits feed directly into item creation, combat mechanics, and skills.

### 2. World Structure Model

The spatial hierarchy of the world is organized as:

$$\text{Continent} \longrightarrow \text{Region} \longrightarrow \text{Biome} \longrightarrow \text{Sub-zone} \longrightarrow \text{Encounter Zone}$$

```yaml id="world-structure-rule"
text: "World structure is hierarchical but non-linear, allowing free traversal while maintaining emergent difficulty scaling based on biome and region."
```

Each tier is a functional container of systemic parameters:
* **Continent**: The ultimate thematic, political, and astronomical container.
* **Region**: Geographically continuous territory (e.g., *Eldret Farmlands*).
* **Biome**: Environmental template determining creature species, climate hazards, and materials.
* **Sub-zone**: High-density hubs or specialized vaults (e.g., *Ymirr's Hidden Cavern*).
* **Encounter Zone**: A micro-area hosting a specific spawn pattern (Solo, Pack, Horde, Boss).

### 3. Biome System

Biomes act as environmental templates modifying core gameplay attributes:

```yaml id="biome-rule"
text: "Biomes define environmental modifiers that influence monsters, loot, and combat conditions without rigidly locking content to specific zones."
```

#### Canonical Biome Matrix

| Biome ID | Name | Elemental Affinity | Key Hazards | Monster Families | Rarity Roll Bonus |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **forest** | Whispering Woodlands | Nature | None | Beast, Plant, Goblin | 1.0x (Standard) |
| **desert** | Shifting Sands | Earth | Heat, Sandstorms | Reptile, Elemental, Undead | 1.1x |
| **frozen_tundra** | Frostbite Wastes | Ice | Hypothermia (Stamina debuff) | Beast, Giant, Undead | 1.15x |
| **swamp** | Mire of Murk | Poison | Toxic Gas Miasma (Poison DoT) | Reptile, Plant, Aberration | 1.2x |
| **volcanic** | Cinder Wastes | Fire | Magma Flows, Ash Choke (Silence) | Demon, Elemental, Dragonkin | 1.3x |
| **sacred_lands** | Sanctuary of Light | Holy | Radiant Burn (Unholy debuff) | Celestial, Construct | 1.4x |
| **corrupted_zones** | The Abyss Scar | Shadow | Sanity Drain, Void Fissures | Undead, Demon, Voidspawn | 1.5x |

### 4. Spawn Ecosystem System

Monster density and spawn rates are completely dynamic, scaling with player activity to create tension.

```yaml id="spawn-ecosystem-rule"
text: "Monster spawns are dynamically generated based on biome density, player activity, and regional encounter balancing systems."
```

* **Adaptive Spawning**: The server adjusts spawn rates on the fly:
  $$\text{SpawnRateSeconds} = \frac{\text{BaseRateSeconds}}{1.0 + (\text{PlayerActivityScore} \times 0.25) + (\text{RegionalDangerFactor} \times 0.15)}$$
* **Escalation**: Higher activity scores shift encounter frequencies from **Solo/Pack** towards **Horde/Boss** compositions, triggering regional defense alerts.

### 5. World Difficulty Model

There are no static, level-gated barriers. Level scaling is soft and emergent, built on tactical monster composition.

```yaml id="difficulty-rule"
text: "World difficulty is emergent and based on encounter composition rather than static level zoning."
```

* **Composition-based Danger**: A level 10 player can enter a high-tier zone, but the presence of coordinated elite squads, heavy environmental hazards, and complex mob synergies will demand mechanical execution or group support rather than a gear check.
* **Elite and Buff Cohesion**: Proximity between monsters of complementary families activates synergy bonuses (e.g., standard physical Goblins receive shielding buffs when spawning near void-infected Goblins).

### 6. Loot Integration Layer

```yaml id="loot-world-rule"
text: "Loot distribution is biome and encounter-driven, with region-specific modifiers influencing item generation and rarity outcomes."
```

* **Affix Scaling**: Items dropped in the *Cinder Wastes* have a heavy weighting bias towards *Strength* and *Fire Damage/Resistance* affixes.
* **Material Harvesting**: Biomes regulate the types of base tier and high-end crafting materials available for extraction, establishing a decentralized global economy.

### 7. Progression Integration

```yaml id="progression-world-rule"
text: "Player progression is not restricted by world zones and is instead shaped by performance-based XP and biome-driven encounters."
```

* **Dynamic Underdog Bonus**: Defeating encounter compositions rated higher than the player's level yields a multiplicative experience bonus:
  $$\text{BonusMultiplier} = \left( \frac{\text{EncounterAverageLevel}}{\text{PlayerLevel}} \right)^{1.5}$$
* **Sustained Chain-Kills**: Maintaining continuous combat across multiple biomes stacks a temporary experience multiplier.

### 8. Travel System

```yaml id="travel-rule"
text: "World traversal is free-form with optional fast travel systems and no level-gated geographic restrictions."
```

* **Unrestricted Travel**: Any level player can walk, ride, or sail anywhere in the world.
* **Special Sub-zones**: Restricted zones (like *Ymirr's Hidden Cavern*) are obscured from standard UI menus, forcing players to discover physical secret passages, solve quests, or scale frozen crags to gain access.
