## WIP (work-in-progress)
  This Program is not meant for use as a module in other go program, Use at your own peril.

## Instalation 
  >[!NOTE]
  >Only LINUX is supported (or any system with uinput + evdev support)\
  >If you want WINDOWS support see Roadmap

  Clone and build:
  ```bash
  git clone https://github.com/ThomasT75/kbdmagic.git
  cd ./kbdmagic/
  ./build.sh 
  ```
  A 'kbdmagic' binary should appear in the current directory 

  Make it executable:
  ```bash
  chmod +x ./kbdmagic
  ```
  Everything in one command:
  ```bash
  git clone https://github.com/ThomasT75/kbdmagic.git && cd ./kbdmagic/ && ./build.sh && chmod +x ./kbdmagic
  ```
  You can delete everything but the `./portable/` directory and the executable `kbdmagic` if you don't want to trinker with the source code

  >[!IMPORTANT] 
  >You will also need to be part of the group `input` sorry

## Features 
  - One Key To One Button Remaps 
  - One Rel To One Axis Remaps
  - Each Axis Direction is Remapable as a Button 
  - Bool Remap: Remaps that apply only on certain Key State (Down/Up)
  - Sequence Remap: Does a Sequence of Inputs in Order
  - Options: See common/options/options.go (Yes RTFM)
  - None Remap: use the keyword none in a bool remap do disable that input

## Usage
  >[!IMPORTANT] 
  >To turn gamepad mode on/off your keyboard/mouse press `PAUSE` (toggle grab) or `DELETE` to quit `kbdmagic`\
  >If your keyboard doesn't have these buttons you will need to edit the source code for another button before building

  While in the same directory as `kbdmagic`:

  Running:
  ```bash
  ./kbdmagic #remap_name 
  ```
  Now to turn On/Off Gamepad mode press the `PAUSE` key in your keyboard.

  You can also select any remap by it's name like: `Default.remap` == `Default`:
  ```bash
  ./kbdmagic Default
  #same result as running ./kbdmagic
  ```
  To list remaps names:
  ```bash
  ./kbdmagic -lr
  ```

  >[!IMPORTANT]
  >In case your mouse or keyboard wasn't picked up and you are in the `input` group\
  >You can use `./kbdmagic -help` and read the description for `-mouse`, `-keyboard` and `-ld` options\
  >Example if your mouse is called "mice" instead: `./kbdmagic -mouse mice`
## Making Remaps
  Make a file in `./portable/remaps/` name it and put a `.remap` extension on it and start editing it:
  > [!CAUTION]
  >This language is sensitive to newlines and is case sensitive.\
  >The compiler will not stop you for coding something that won't work / make sense.\
  >Also the compiler is being rewriten, that is why.

### Remap Notes: 
  ```go 
  // An KEY or BTN can only be bound to a BUTTON
  // An REL will only act like a KEY or BTN if it has a -/+ sign as a suffix
  // An REL can only be bound to a AXIS
  // You can add a FORCE multiplier to any AXIS
  
  // If you Remap the same KEY or REL or BTN only the last one will take effect
  // but will cause an error in future versions

  // You can add the prefIx "toggle" to a KEY or BTN and will toggle the remap
  // instead of following the KEY or BTN state
  // bool remap and normal remap can have "toggle" prefix 
  
  // You can use any controller naming scheme to map to 
  // (will still be a generic controller in the end)
  // See ecodes/ecodes.go for all exact names
  ```
> [!TIP]
> You can comment a line using "//" and everything after it will not be read by the compiler
### Keyboard Example:
  ```go
  // This will press the A button every time 
  // you press the Space bar on your keyboard 
  KEY_SPACE : GP_BTN_A

  // Map WASD to the left analog axis
  KEY_W : GP_AXIS_Y-
  KEY_A : GP_AXIS_X-
  KEY_S : GP_AXIS_Y+
  KEY_D : GP_AXIS_X+
  ```
  >[!TIP]
  >You don't need spaces on either side of ":" ex: "KEY_SPACE:GP_BTN_A" also works
### Mouse Example:
  ```go
  // Example of using different controller naming schemes
  BTN_LEFT : GP_BTN_R2 
  BTN_RIGHT : GP_BTN_ZL

  // Mouse Movement To Right Analog
  REL_X : GP_AXIS_RX
  REL_Y : GP_AXIS_RY

  // Each direction of the Mouse wheel is a different button click
  REL_WHEEL+ : GP_BTN_R1
  REL_WHEEL- : GP_BTN_L1
  ```
### Bool Remap Example:
  ```go 
  // Can be bound to any KEY or BTN

  // if the left alt key is OFF (aka: UP)
  // the left analog will go to 100% of the way when you press WASD
  // if the left alt key is ON (aka: DOWN)
  // the left analog will only go to 50% of the way when you press WASD
  KEY_LEFTALT {
    OFF 
    KEY_W : GP_AXIS_Y-
    KEY_A : GP_AXIS_X-
    KEY_S : GP_AXIS_Y+
    KEY_D : GP_AXIS_X+
    ON 
    KEY_W : GP_AXIS_Y- 0.5
    KEY_A : GP_AXIS_X- 0.5
    KEY_S : GP_AXIS_Y+ 0.5
    KEY_D : GP_AXIS_X+ 0.5
  }

  // You can still remap the left alt to a button while being a bool remap
  KEY_LEFTALT : GP_BTN_L3 // crouch ?

  // Now you can imagine this remap is used for walking slowly while crouching
  // You can use this remap function to do a remap with
  // block, parry, attack only using 2 keys or mouse buttons

  // If you want to only update the remap on the ON or OFF state 
  // you can just omit the other state

  // You can unbind a key using the "none" keyword instead of a gamepad button

  // You cannot put a bool remap inside a bool remap 
  
  // Putting an empty line inside a bool remap is not accepted by the program 
  ```

### Sequence 

[//]: # "
// The best example I can give you is a fighting game string 
// So lets get the most op combo for dizzy a character in GGACR+
KEY_1 : [GP_BTN_RT, GP_AXIS_Y+(750ms), GP_AXIS_Y-, GP_AXIS_X-,
GP_AXIS_X+, GP_AXIS_Y+, GP_BTN_B + GP_BTN_X])
"

```go 
  // Can be bound to any KEY or REL with +/- sign  
  
  // A sequence will run each command in order 
  // right after releasing the previous command

  // Reload in MGS2
  KEY_R : [GP_BTN_R2, GP_BTN_R2]

  // You can also specify the type of operation (default is click) 
  // hold/release

  // You can also specify a delay between commands

  // The orignal plan for this was to make something like this for mhgu items
  REL_WHEEL- : [hold GP_BTN_L1, 300ms, GP_BTN_NS_A, 100ms, release GP_BTN_L1]
  REL_WHEEL+ : [hold GP_BTN_L1, 300ms, GP_BTN_NS_Y, 100ms, release GP_BTN_L1]
  // But if you use this you will notice that it sucks for this use case

  // lets say you want to do the konami code 
  // every time you press the space bar
  KEY_SPACE : [GP_DPAD_UP, GP_DPAD_UP, GP_DPAD_DOWN, GP_DPAD_DOWN, GP_DPAD_LEFT,
    GP_DPAD_RIGHT, GP_DPAD_LEFT, GP_DPAD_RIGHT, GP_DPAD_NS_B, GP_DPAD_NS_A]

  // You can place a newline after a comma 
  
  // Now lets say you want to do the konami code again 
  // but hold the first key for 1 sec and click b + a together 
  KEY_SPACE : [GP_DPAD_UP(1000ms), GP_DPAD_UP, GP_DPAD_DOWN, GP_DPAD_DOWN,
    GP_DPAD_LEFT, GP_DPAD_RIGHT, GP_DPAD_LEFT, GP_DPAD_RIGHT,
    GP_DPAD_NS_B + GP_DPAD_NS_A]

  // The limit of a sequence is 256 outputs 
  // Note that limit is shared with the queue meaning
  // you can only queue 1 sequence of 256 outputs 
  // 2 sequences of 128 outputs and so on 
  // I don't expect you to run into this problem 
  // but who knows there might be that one guy 
  // trying to beat a whole game using 1 sequence
```
### Options
```go 
  // Options yay 
  !SomeNerdOptionName = SomeNerdOptionValue
  // Check common/options/options.go for a list of options and they meaning
  // Options should appear before any remap
```

  If you open `./grammar.txt` you can see an more complex overview of the remap language it is not exactly how the program will interpret but it should give you an idea of how to code in this remap language

  I'm not writing an full doc for this language yet (haven't found good solutions for offline viewing)

## Roadmap
  - [ ] GUI (using fyne)
    - This will come in the form of another repo that uses this repo for the actual functionally of the GUI. Which means a lot of work to get this repo to that standard
  - [ ] More remap Language features
    - [x] Bool Remap 
    - [x] Sequence 
    - [x] Options
    - [x] Toggle for K&M inputs
    - [x] "none" Keyword
    - [ ] Turbo 
    - [ ] Combo
    - [ ] "Macros" 
      - A way more complicated version of Sequence
    - [ ] Passthrough 
    - [ ] Key to Key
    - [ ] Ramp 
      - The time you hold a key translates into the axis position
    - [ ] Layers 
      - Like Bool Remap but can de/activate with any key and it gets exclusivity while active
  - [ ] Misc Features
    - [ ] Vibration as sound
    - [ ] CLI option for hot reloading current remap file
    - [ ] Replug virtual controller with Keybind
  - [ ] Expose at runtime some of the configuration that is done at compile time
    - [x] Support for Remap Options 
    - [ ] Support for CLI Options like `PAUSE` & `DELETE` 
  - [ ] Controller Emulation Support 
    - [x] Generic
    - [ ] DS4
      - [ ] Controller
      - [ ] Touchpad
      - [ ] Gyro
    - [ ] X360
    - [yannbouteiller/vgamepad](https://github.com/yannbouteiller/vgamepad/tree/main/vgamepad/lin)
      has an example of the implementation in python for both DS4 and X360 for LINUX
  - [ ] Misc
    - [ ] make install script for udev rule to remove the input group requirement  
  - [ ] WINDOWS Support Maybe?
    - [ ] Learn go-hook 
      - 4 years without a update \:)
    - [ ] Look into ViGEmBus driver 
      - python implementation by [yannbouteiller/vgamepad](https://github.com/yannbouteiller/vgamepad/tree/main/vgamepad/win)
      - golang implementation by [openstadia/go-vigem](https://github.com/openstadia/go-vigem)\
      it is just a warper around a the ViGEmBus .dll

## Contrib
  Don't \:)

[//]: # "But if you want, fork it and toy with it 
It is very much a personal project, 
I have more fun that way as I can do whatever,
and I want you to have the same experience.
"
