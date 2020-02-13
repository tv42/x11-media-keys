# x11-media-keys -- Adjust volume and screen brightness based on keypresses

Current status: Wayland won. This should still keep working ok, but if
it doesn't, I might not fix it anymore.

Mainstream Linux desktop integration is reaching a point where using
an alternate window manager causes all sorts of annoyances. With
Ubuntu 15.04, the panel refuses to work with xmonad for me, so I'm
replacing it with something simpler.

`x11-media-keys` listens for volume up/down/mute and brightness
up/down keys and does the appropriate actions via ALSA and the X11
RandR backlight property.
