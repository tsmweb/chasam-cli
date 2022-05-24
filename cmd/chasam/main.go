package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/tsmweb/chasam/app/fstream"
	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/app/media"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

var (
	cpu      = flag.Int("cpu", runtime.NumCPU(), "--cpu=4")
	source   = flag.String("source", "", "--source=image/source")
	target   = flag.String("target", "", "--target=image/target")
	hashType = flag.String("hash", "d-hash", "--hash=sha1,ed2k,a-hash,d-hash,d-hash-v,p-hash")
	distance = flag.Int("distance", 10, "--distance=10")

	countMatchCh = make(chan struct{})
)

func main() {
	flag.Parse()
	start := time.Now()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func(ctx context.Context, fn context.CancelFunc) {
		<-ctx.Done()
		fn()
	}(ctx, stop)

	if *source == "" || *target == "" {
		printHelper()
		os.Exit(0)
	}

	countMatch := 0
	go func() {
		for range countMatchCh {
			countMatch++
		}
	}()

	if err := runMediaSearchStream(ctx); err != nil {
		fmt.Printf("[!] Error: %v\n", err.Error())
	}

	elapsed := time.Since(start)
	fmt.Printf("\n[>] Pesquisa concluída em: %s\n", elapsed)
	fmt.Printf("[>] Total de arquivos encontrados: %d\n", countMatch)
	time.Sleep(time.Second)
}

func runMediaSearchStream(ctx context.Context) error {
	roots := strings.Split(*target, ",")
	hashTypes := makeHashTypes()

	hashStorage, err := hash.NewStorage(*source, hashTypes)
	if err != nil {
		return err
	}

	fstream.NewMediaSearchStream(ctx, roots, *cpu).
		OnError(func(err error) {
			fmt.Printf("[!] Error: %v\n", err.Error())
		}).
		OnEach(func(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Filter image
			if m.Type() == "image" {
				return fstream.Next, nil
			}
			return fstream.Skip, nil
		}).
		OnEach(func(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by SHA1
			h, err := m.SHA1()
			if err != nil {
				return fstream.Skip, err
			}
			if src := hashStorage.FindByHash(hash.SHA1, h); src != "-1" {
				printMatch(src, m.Name(), m.Path(), "SHA1", 0)
				return fstream.Match, nil
			}
			return fstream.Next, nil
		}).
		OnEach(func(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by ED2K
			h, err := m.ED2K()
			if err != nil {
				return fstream.Skip, err
			}
			if src := hashStorage.FindByHash(hash.ED2K, h); src != "-1" {
				printMatch(src, m.Name(), m.Path(), "ED2K", 0)
				return fstream.Match, nil
			}
			return fstream.Next, nil
		}).
		OnEach(func(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by AHash
			h, err := m.AHash()
			if err != nil {
				return fstream.Skip, err
			}
			if dist, src := hashStorage.FindByPerceptualHash(hash.AHash, h[0], *distance); dist != -1 {
				printMatch(src, m.Name(), m.Path(), "AHash", dist)
				return fstream.Match, nil
			}
			return fstream.Next, nil
		}).
		OnEach(func(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by DHash
			h, err := m.DHash()
			if err != nil {
				return fstream.Skip, err
			}
			if dist, src := hashStorage.FindByPerceptualHash(hash.DHash, h[0], *distance); dist != -1 {
				printMatch(src, m.Name(), m.Path(), "DHash", dist)
				return fstream.Match, nil
			}
			return fstream.Next, nil
		}).
		OnEach(func(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by DHashV
			h, err := m.DHashV()
			if err != nil {
				return fstream.Skip, err
			}
			if dist, src := hashStorage.FindByPerceptualHash(hash.DHashV, h[0], *distance); dist != -1 {
				printMatch(src, m.Name(), m.Path(), "DHashV", dist)
				return fstream.Match, nil
			}
			return fstream.Next, nil
		}).
		OnEach(func(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by PHash
			h, err := m.PHash()
			if err != nil {
				return fstream.Skip, err
			}
			if dist, src := hashStorage.FindByPerceptualHash(hash.PHash, h[0], *distance); dist != -1 {
				printMatch(src, m.Name(), m.Path(), "PHash", dist)
				return fstream.Match, nil
			}
			return fstream.Next, nil
		}).
		OnMatch(func(m *media.Media) {
			countMatchCh <- struct{}{}
		}).
		Run()

	close(countMatchCh)
	return nil
}

func makeHashTypes() []hash.Type {
	var hashTypes []hash.Type
	_hashTypes := strings.Split(*hashType, ",")

	for _, ht := range _hashTypes {
		switch strings.ToLower(ht) {
		case "sha1":
			hashTypes = append(hashTypes, hash.SHA1)
		case "ed2k":
			hashTypes = append(hashTypes, hash.ED2K)
		case "a-hash":
			hashTypes = append(hashTypes, hash.AHash)
		case "d-hash":
			hashTypes = append(hashTypes, hash.DHash)
		case "d-hash-v":
			hashTypes = append(hashTypes, hash.DHashV)
		case "p-hash":
			hashTypes = append(hashTypes, hash.PHash)
		case "w-hash":
			hashTypes = append(hashTypes, hash.WHash)
		}
	}

	return hashTypes
}

func printMatch(sourceName, targetName, targetPath, hashType string, distHamming int) {
	fmt.Printf("\n[*] Origem: %s\n", sourceName)
	fmt.Printf("\t- Match: %s\n", targetName)
	fmt.Printf("\t- Path: %s\n", targetPath)
	fmt.Printf("\t- Tipo do hash: %s\n", hashType)
	fmt.Printf("\t- Distância de hamming: %d\n", distHamming)
}

const templateHelperStr = "\t%-10s \t\t %s\n"

func printHelper() {
	fmt.Println("Uso: chasam --source=images/source --target=images/target-1,images/target-2 --hash=d-hash,d-hash-v --distance=10")
	fmt.Println("Realiza uma pesquisa de imagens através da comparação de hashs.")

	fmt.Printf("\nArgumentos.\n")
	fmt.Printf(templateHelperStr, "--cpu", "defini o número de núcleos da cpu para o processamento dos hash")
	fmt.Printf(templateHelperStr, "--distance", "distância limite entre dois hashs perceptivos")
	fmt.Printf(templateHelperStr, "--hash", "tipo do hash "+
		"(pode ser informado mais de um tipo separados por vírgula)")

	fmt.Printf(templateHelperStr, "\tsha1", "função hash criptográfica de 160 bits")
	fmt.Printf(templateHelperStr, "\ted2k", "hash usado em compartilhamento de arquivos eDonkey")

	fmt.Printf(templateHelperStr, "\ta-hash", "hash médio "+
		"(calculado pela média de todos os valores de cinza da imagem)")

	fmt.Printf(templateHelperStr, "\td-hash", "hash de diferença "+
		"(calcula a diferença entre um pixel e seu vizinho da direita, seguindo o degrade horizontal)")

	fmt.Printf(templateHelperStr, "\td-hash-v", "hash de diferença vertical "+
		"(calcula a diferença entre um pixel se seu vizinho abaixo, seguindo o degrade vertical)")

	fmt.Printf(templateHelperStr, "\tp-hash", "perceptual hash (calcula aplicando uma transformada discreta de cosseno)")

	fmt.Printf(templateHelperStr, "\tw-hash", "wavelet hash (calcula aplicando uma transformada wavelet bidimensional)")

	fmt.Printf(templateHelperStr, "--source", "diretório de origem com as imagens/vídeos a serem pesquisadas")

	fmt.Printf(templateHelperStr, "--target", "diretório alvo onde será realizada a pesquisa por imagens/vídeos "+
		"(pode ser informado mais de um diretório separados por vírgula)")
}
