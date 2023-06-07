
# How to use

## API有効化
gcloud services enable secretmanager.googleapis.com --project YOUR_PROJECT

## サービスアカウントに権限追加
gcloud projects add-iam-policy-binding YOUR_PROJECT --member serviceAccount:YOUR_PROJECT@appspot.gserviceaccount.com --role roles/secretmanager.secretAccessor

## デプロイ
gcloud functions deploy notify \
		--trigger-http \
		--allow-unauthenticated \
		--runtime go120 \
		--env-vars-file env.yaml \
		--region asia-northeast1 \
		--project YOUR_PROJECT
