
# How to use

## API有効化
gcloud services enable secretmanager.googleapis.com --project YOUR_PROJECT

## サービスアカウントに権限追加
gcloud projects add-iam-policy-binding YOUR_PROJECT --member serviceAccount:YOUR_PROJECT@appspot.gserviceaccount.com --role roles/secretmanager.secretAccessor

## デプロイ
gcloud functions deploy notify-v2 \
		--gen2 \
		--trigger-http \
		--allow-unauthenticated \
		--entry-point Notify \
		--runtime go124 \
		--env-vars-file env.yaml \
		--region asia-northeast1 \
		--project mogily-goods-poc \
		--source .
