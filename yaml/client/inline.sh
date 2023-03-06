cat <<EOF | kubectl create -oyaml -f -
apiVersion: cbt.storage.k8s.io/v1alpha1
kind: VolumeSnapshotDeltaToken
metadata:
  name: test
spec:
  baseVolumeSnapshotName: vs-00
  targetVolumeSnapshotName: vs-01
  mode: block
EOF
