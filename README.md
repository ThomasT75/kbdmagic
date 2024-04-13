## WIP (work-in-progress)
  Everything in this repo can change without a CARE for backwards compatibility 

## Instalation 
  >[!NOTE]
  >Only LINUX is supported (or any system with uinput + evdev support)\
  >If you want WINDOWS support just open a issue telling which go modules to replace uinput and evdev with on WINDOWS

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
## Usage
  >[!CAUTION] 
  >This program will grab your mouse and keyboard as soon as you run the `./kbdmagic`\
  >To release your keyboard and mouse press `PAUSE` (ungrabs/regrabs) or `DELETE` (quits `kbdmagic`)\
  >If your keyboard doesn't have these buttons you will need to unplug and replug your keyboard or mouse to crash `kbdmagic` or edit the source code for another button before building

  While in the same directory as `kbdmagic`:\
  Run `./kbdmagic` and it will look at the `./portable/remap/` directory in the same place as the executable for a remap by default will pick `Default.remap`
  ```bash
  ./kbdmagic
  ```
  You can also select any remap by it's name like: `Default.remap` == `Default` 
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
  >Example if you mouse is called "mice" instead: `./kbdmagic -mouse mice`
## Making Remaps
  Make a file in `./portable/remaps/` name it and put a .remap extension on it\
  Next you will code the remap very simple just open the file you just created in a text editor and follow the example:
  ```go
  // Keyboard example
  // This will press the A button on a generic controller every time you press the Space bar on your keyboard 
  KEY_SPACE : GP_BTN_A

  // Map WASD to the left analog axis
  KEY_W : GP_AXIS_Y-
  KEY_A : GP_AXIS_X-
  KEY_S : GP_AXIS_Y+
  KEY_D : GP_AXIS_X+

  // Mouse example
  // you can use any controller naming scheme to map to (will still be a generic controller in the end)
  // see ecodes/ecodes.go for all exact names like GP_BTN_NS_A for the A button in a NS
  BTN_LEFT : GP_BTN_R2 
  BTN_RIGHT : GP_BTN_ZL

  // Mouse Movement To Right Analog
  REL_X : GP_AXIS_RX
  REL_Y : GP_AXIS_RY
 
  ```
  >[!TIP]
  >You don't need spaces between ":" ex: "KEY_SPACE:GP_BTN_A" also works

  All you can do right now with the remaps is 1 to 1 translation (incluing axis)\
  If you open `./portable/remaps/Default.reamp` you can see more examples but those are not documented


## Roadmap
  - [ ] GUI (using fyne)
  - [ ] More complex remap Language features
  - [ ] Expose at runtime some of the configuration that is done at compile time
  - [ ] WINDOWS Support Maybe?
