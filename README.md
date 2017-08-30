# sudolikeaboss - Now you can too!

**NOTE**: This is a fork from https://github.com/ravenac95/sudolikeaboss.

![Demo](https://raw.githubusercontent.com/ravenac95/readme-images/master/sudolikeaboss/demo.gif)

Pretty neat, eh?

## What's happening here?

`sudolikeaboss` is a simple application that aims to make your life as a dev, ops, or just a random person who likes to ssh and sudo into boxes much, much easier by allowing you to access your `1Password` passwords on the terminal.
All you need is [`iterm2`](https://iterm2.com/), [`1Password`](https://agilebits.com/onepassword), a mac, and a dream.

## Benefits

- Better security through use of longer, more difficult to guess passwords
- Better security now that you can have a different password for every server if you'd like
- Greater convenience when accessing passwords on the terminal


## So is this only for sudo passwords?

No! You can use this for tons of things! Like...

- [`dm-crypt`](https://code.google.com/p/cryptsetup/wiki/DMCrypt) passwords on external boxes
- [`gpg`](https://www.gnupg.org/) passwords to use on the terminal

## Ok! I want it. How do I install this thing?!

I tried to make installation as simple as possible. So here's the quickest path to awesomeness.

### Install with homebrew

*This is by far the easiest method, and the one I recommend most.*

```
$ brew install brycekahle/sudolikeaboss/sudolikeaboss
```

### Install from source

Assuming that you have Go installed and you know how to use it's associated tools...

``` 
$ go get github.com/brycekahle/sudolikeaboss
```

The `sudolikeaboss` binary should now be in `$GOPATH/bin/sudolikeaboss`

### Install from zip

Download one of the following zips:

- amd64: https://github.com/brycekahle/sudolikeaboss/releases/download/v0.3.0/sudolikeaboss_v0.3.0_darwin_amd64.zip

*warning*: At this time I'm not sure if the 386 version works. In theory it should, but I don't have access to a 32-bit machine to test this.

Then, unzip the file and copy it to the desired location for installation (I suggest `/usr/local/bin/sudolikeaboss`).

This entire workflow, would look like this::

```
$ mkdir sudolikeaboss
$ cd sudolikeaboss
$ wget https://github.com/brycekahle/sudolikeaboss/releases/download/v0.3.0/sudolikeaboss_v0.3.0_darwin_amd64.zip
$ unzip sudolikeaboss_v0.3.0_darwin_amd64.zip
$ cp sudolikeaboss /usr/local/bin/sudolikeaboss
```

## Configure `iterm2` to use `sudolikeaboss`

After installing `sudolikeaboss`, you still need to configure `iterm2`. This is fairly simple. Just watch this gif!

![Configuration](https://raw.githubusercontent.com/ravenac95/readme-images/master/sudolikeaboss/configuration.gif)


## Configuring 1Password5 to work with sudolikeaboss

If you're using 1Password5, or you run into this screen:

![1Password5 Error](https://raw.githubusercontent.com/ravenac95/readme-images/master/sudolikeaboss/cannot-fill-item-error-popup.png)

This causes a problem for `sudolikeaboss` as it isn't a "trusted browser" per se. In order to fix this issue, you need to change some preferences on your 1Password installation. Open up 1Password's preferences and find the 
`Advanced` settings tab. Then make sure to uncheck the option 
`Verify browser code signature`. After doing that, `sudolikeaboss` 
should work... like a boss. For the visual learners here's a screenshot:

![1Password Config Change](https://cloud.githubusercontent.com/assets/889219/6270365/a69a0726-b816-11e4-9b96-558ddeb00378.png)

## Getting passwords into 1Password

To get 1Password to play ball, just make sure that any passwords you set use `sudolikeaboss://local` as the website on the 1Password UI. Watch this example:

![Add Password Demo](https://raw.githubusercontent.com/ravenac95/readme-images/master/sudolikeaboss/add-password.gif)

## Potential Plans for the future!

These are some ideas I have for the future. This isn't an exhaustive list, and, more importantly, I make no guarantees on whether or not I can or will get to any of these.

- Ability to save passwords directly from the command line. Of any of these plans, this is probably the most feasible. Again, no promises, but I personally want this feature too
- ``tmux`` support. So for those of you that don't use iterm2 I may be able to create a different kind of plugin that can work with this.
- linux support? This is a big question mark. If I can get tmux support to work, then presumably doing something similar for linux wouldn't be impossible. However, the other hard part of this is that linux doesn't currently have a GUI for 1Password, but I actually have plans to attempt to create a gui using some already built tools.

## Gotchas/Known Issues

Here are just some questions or gotchas that I figured people would run into or have.

### Why is the 1Password popup not where I'm typing?

The way the popup works is by finding your mouse cursor. I'd like to improve this, but since I'm using 1Password's undocumented API this is how it will be right now.

### I don't use 1Password

Are you serious?! If you're on a mac and you have passwords, you should be using 1Password. With that said, I would love to support additional password managers as the project grows. 

### I use linux

Sorry :( I don't have anything for you yet. Maybe you can help me with that :)

### I use Windows

Unfortunately, I have no current plans to do this on Windows. This is mostly because I wouldn't know where to start. At the moment this software is pretty dependent on somethings like iterm2 and 1Password. As my expertise is in Linux/Unix environments and not in Windows, I'm not exactly sure what tools/workflow someone in that camp would use. If you'd like to help out in this arena, I would be more than happy to give it all a shot.

### What's that weird icon on the top-right of the iterm2 window?

That's just an icon that indicates that an iterm2 [coprocess](https://iterm2.com/coprocesses.html#/section/home) is running. It
will disappear eventually, as `sudolikeaboss` times out after 30 seconds when waiting for user input.

### Do you have this "undocumented API" documented somewhere?

Not yet, but it will happen soon, hopefully.

## Contributing/Developing

Coming soon.
