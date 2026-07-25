#!/usr/bin/env bash
# Switch back to the GUI desktop (restart LightDM to re-acquire DRM master on Pi 5)
export WLR_DRM_DEVICES=/dev/dri/card1:/dev/dri/card0
sudo systemctl restart lightdm
