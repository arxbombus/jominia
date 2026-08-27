# Jomini Syntax

## Scalars

```jomini
aaa=foo         # a plain scalar
bbb=-1          # an integer scalar
ccc=1.000       # a decimal scalar
ddd=yes         # a true scalar
eee=no          # a false scalar
fff="foo"       # a quoted scalar
ggg=1821.1.1    # a date scalar in Y.M.D format
-1=aaa
"1821.1.1"=bbb
@my_var="ccc"    # define a variable
a=1 b=2 c=3 # One can have multiple key values pairs per line as long as boundary character is separating them
hhh="a\"b"      # escaped quote. Equivalent to `a"b`
iii="\\"        # escaped escape. Equivalent to `\`
mmm="\\\""      # multiple escapes. Equivalent to `\"`

# a multiline quoted scalar
ooo="hello
     world"

# Quotes can contain escape codes! Imperator uses them as
# color codes (somehow `0x15` is translated to `#` in the
# parsing process)
nnn="ab <0x15>D ( ID: 691 )<0x15>!"
```

## Arrays / Objects

```jomini
flags={
    schools_initiated=1444.11.11
    mol_polish_march=1444.12.4
}
players_countries={
    "Player"
    "ENG"
}
campaign_stats={ {
    id=0
    comparison=1
    key="game_country"
    selector="ENG"
    localization="England"
} {
    id=1
    comparison=2
    key="longest_reign"
    localization="Henry VI"
} }
```

## Operators

```jomini
intrigue >= high_skill_rating
age > 16
count < 2
scope:attacker.primary_title.tier <= tier_county
a != b
start_date == 1066.9.15
c:RUS ?= this
```

## Boundary Characters

- Whitespace
- Open ({) and close (}) braces
- An operator
- Quotes
- Comments

`a={b="1"c=d}foo=bar#good` is equivalent to:

```jomini
a = {
  b = "1"
  c = d
}
foo = bar # good
```

## Comments

```jomini
my_obj = # this is going to be great
{ # my_key = prev_value
    my_key = value # better_value
    a = "not # a comment"
} # the end
```

## Other Cases

An object / array value does not need to be prefixed with an operator:

```jomini
foo{bar=qux}

# is equivalent to `foo={bar=qux}`
```

A value of {} could mean an empty array or empty object depending on the context. We would probably need `.cwt` config rules for this (from which we can generate/build our own rules).

```jomini
discovered_by={}
```

Any number of empty objects / arrays can occur in an object and should be skipped.

```jomini
history={{} {} 1629.11.10={core=AAA}}
```

An object can be both an array and an object at the same time:

```
brittany_area = { #5
    color = { 118  99  151 }
    169 170 171 172 4384
}
```

The previous example showed how an object transitions to an array as seen in EU4 game files. In CK3 there is the opposite occurrence as shown below: an array transitions to an object. I colloquially refer to these as array trailers (EU4) and hidden objects (CK3):

```jomini
levels={ 10 0=2 1=2 }
# I view it as equivalent to
# levels={ { 10 } { 0=2 1=2 } }
```

Scalars can have non-alphanumeric characters:

```jomini
flavor_tur.8=yes
dashed-identifier=yes
province_id = event_target:agenda_province
@planet_standard_scale = 11
```

Variables can be used in interpolated expressions:

```jomini
position_x = @[1-leo_x]
```

Don’t try to blank store all numbers as 64 bit floating point, as there are some 64 bit unsigned integers that would cause floating point to lose precision:

```jomini
identity=18446744073709547616

# converted to floating point would equal:
# identity=18446744073709548000
```

Equivalent quoted and unquoted scalars are not always intepretted the same

```jomini
unit_type="western"  # bad: save corruption
unit_type=western    # good
```

The type of an object or array can be externally tagged:

```jomini
color = rgb { 100 200 150 }
color = hsv { 0.43 0.86 0.61 }
color = hsv360{ 25 75 63 }
color = hex { aabbccdd }
mild_winter = LIST { 3700 3701 }
```

The EU4 1.26 (Dharma) patch introduced parameter syntax that hasn’t been seen in other PDS titles. From the changelog:

```jomini
generate_advisor = {
  [[scaled_skill]
    $scaled_skill$
  ]
  [[!skill] if = {} ]
}
```

Semi-colons at the end of quotes (potentially lines) are ignored.

```jomini
textureFile3 = "gfx//mapitems//trade_terrain.dds";
```

There are unmarked lists in CK3 and Imperator. Typically lists are use brackets ({, }) but those are conspicuously missing here:

```jomini
simple_cross_flag = {
  pattern = list "christian_emblems_list"
  color1 = list "normal_colors"
}
```

Alternating value and key value pairs. Makes one wish they used a bit more of a self describing format:

```jomini
on_actions = {
  faith_holy_order_land_acquisition_pulse
  delay = { days = { 5 10 }}
  faith_heresy_events_pulse
  delay = { days = { 15 20 }}
  faith_fervor_events_pulse
}
```

ck3 scripted effect:

```jomini
give_gold = {
   $GIVER$ = { save_scope_as = giver }
   $TAKER$ = { save_scope_as = taker }
   save_scope_value_as = {
      name = gold_amount
      value = $VALUE$
   }
   scope:giver = {
      remove_short_term_gold = scope:gold_amount
      scope:taker = {
         add_gold = scope:gold_amount
      }
   }
   clear_saved_scope = giver
   clear_saved_scope = taker
}
```

Parameters in vic3:

```jomini
transfer_state = { # Changes ownership of any state within a STATE region (e.g. s:STATE_MONTENEGRO) from GIVER to TAKER countries (e.g. c:TUR). Automatically checks GIVER exists and owns the state.
    if = {
        limit = {
            exists = $GIVER$
            exists = $TAKER$
            $TAKER$ != $GIVER$
            $STATE$ = {
                any_scope_state = {
                    $GIVER$ = owner
                }
            }
        }
        $STATE$ = {
            every_scope_state = {
                limit = {
                    owner = $GIVER$
                }
                set_state_owner = $TAKER$
            }
        }
    }
}
transfer_state = {
    STATE = s:STATE_MONTENEGRO
    GIVER = c:TUR
    TAKER = root
}
```

Vic3 gui function macros

```jomini
macro = {
    description = "Add a loc key with a trailing newline if a condition is satisfied"
    definition = "MakeLineIf(Condition, LocKey)"
    replace_with = "ConcatIfNeitherEmpty(AddLocalizationIf(Condition, LocKey), Localize( 'NEWLINE' ))"
}
```

This macro can be called with [MakeLineIf(boolean, loc key)]. For example:

```jomini
[MakeLineIf( IsZero(State.GetTradeCapacity), 'NO_WORLD_MARKET_ACCESS_DUE_TO_NO_TRADE_CAPACITY')]
```

This is equivalent to and read by the game as:

```jomini
[ConcatIfNeitherEmpty(AddLocalizationIf( IsZero(State.GetTradeCapacity), 'NO_WORLD_MARKET_ACCESS_DUE_TO_NO_TRADE_CAPACITY'), Localize( 'NEWLINE' )]
```

@values

```jomini
@pi = 3.1416
@third = @[1/3]
@sub_YUC_hoist_scale = @[third*2]
@canton_scale_cross_x = @[ ( 333 / 768 ) + 0.001 ]
@birthrate_at_growth_max = @[(pop_growth_max_sol-pop_growth_transition_sol)*((min_birthrate-birthrate_at_transition)/(pop_growth_stable_sol-pop_growth_transition_sol))+birthrate_at_transition]
@this_is_you = "this_is_you.dds"
@default_window_file = "gui/notifications/jomini_message.gui"
@default_window_name = "jomini_message"
@tex_star_position = @sixth
```

Vic 3 gui

```jomini
types wargoal_types
{
 type add_wargoal_panel = default_block_window  {
  name = "add_wargoal_panel"
  datacontext = "[AddWarGoalPanel.AccessDiplomaticPlay]"

  blockoverride "window_header_name" {
   text = "ADD_WARGOAL_HEADER"
  }

  blockoverride "entire_back_button" {
   back_button_large = {
    position = { 8 30 }
    onclick = "[AddWarGoalPanel.ClearSelectedWarGoalType]"
    input_action = "back"
    visible = "[AddWarGoalPanel.HasSelectedWarGoalType]"
   }

   back_button_large = {
    position = { 8 30 }
    onclick = "[InformationPanelBar.OpenPreviousPanel]"
    input_action = "back"
    visible = "[Not(AddWarGoalPanel.HasSelectedWarGoalType)]"
   }
  }
 }
}
```

stellaris' `optimize_memory`

```jomini
has_overlord_dlc = {
    optimize_memory
    host_has_dlc = "Overlord"
}
```

I don’t expect any parser to be able to handle all these edge cases in an ergonomic and performant manner.
