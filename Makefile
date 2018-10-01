all: txprocessor blockprocessor eventprocessor utxolister tester

txprocessor:
	docker build -f Dockerfile.txprocessor -t thematterio/plasma:txprocessor -t txprocessor . 
	docker push thematterio/plasma:txprocessor

blockprocessor:
	docker build -f Dockerfile.blockprocessor -t thematterio/plasma:blockprocessor -t blockprocessor . 
	docker push thematterio/plasma:blockprocessor

eventprocessor:
	docker build -f Dockerfile.eventprocessor -t thematterio/plasma:eventprocessor -t eventprocessor . 
	docker push thematterio/plasma:eventprocessor

utxolister:
	docker build -f Dockerfile.utxolister -t thematterio/plasma:utxolister -t utxolister . 
	docker push thematterio/plasma:utxolister

tester:
	docker build -f Dockerfile.tester -t thematterio/plasma:tester -t tester . 
	docker push thematterio/plasma:tester
