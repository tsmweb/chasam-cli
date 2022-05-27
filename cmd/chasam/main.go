package main

/*
go run main.go --source=/home/martins/Desenvolvimento/SPTC/files/source --target=/home/martins/Desenvolvimento/SPTC/files/benchmark/rotate --hash=p-hash --distance=10
*/

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"github.com/tsmweb/chasam/app/fstream"
	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/pkg/progressbar"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	cpu      = flag.Int("cpu", runtime.NumCPU(), "--cpu=4")
	source   = flag.String("source", "", "--source=image/source")
	target   = flag.String("target", "", "--target=image/target")
	hashType = flag.String("hash", "d-hash", "--hash=sha1,ed2k,a-hash,d-hash,d-hash-v,p-hash")
	distance = flag.Int("distance", 10, "--distance=10")

	hashStorage  *hash.Storage
	countFileCh  = make(chan struct{})
	countMatchCh = make(chan struct{})

	_csv *csv.Writer
)

func main() {
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func(ctx context.Context, fn context.CancelFunc) {
		<-ctx.Done()
		fn()
	}(ctx, stop)

	//ctx, cancelFun := context.WithCancel(context.Background())
	//go func() {
	//	os.Stdin.Read(make([]byte, 1)) // read a single byte
	//	cancelFun()
	//}()

	if *source == "" || *target == "" {
		printHelper()
		os.Exit(0)
	}

	csvFile, err := os.Create("match.csv")
	if err != nil {
		fmt.Printf("[!] Error: %v\n", err.Error())
	}
	_csv = csv.NewWriter(csvFile)
	defer func() {
		_csv.Flush()
		csvFile.Close()
	}()
	_csv.Write([]string{"ORIGEM", "ALVO", "ALVO PATH", "TIPO DO HASH", "DISTANCIA DE HAMMING"})

	printBanner()

	var bar progressbar.Bar
	bar.NewOption("=")

	countFile := 0
	go func() {
		for range countFileCh {
			countFile++
			bar.Play(int64(countFile))
		}

		bar.Finish()
	}()

	countMatch := 0
	go func() {
		for range countMatchCh {
			countMatch++
			bar.Match()
		}
	}()

	start := time.Now()

	if err := runMediaSearchStream(ctx); err != nil {
		fmt.Printf("[!] Error: %v\n", err.Error())
	}

	elapsed := time.Since(start)
	time.Sleep(time.Millisecond * 500)

	color.Printf("\n[>] Pesquisa concluída em: <green>%s</>\n", elapsed)
	color.Printf("[>] Total de arquivos analisados: <green>%d</>\n", countFile)
	color.Printf("[>] Total de match: <green>%d</>\n", countMatch)
	color.Printf("[>] Arquivo de match: <green>%s</>\n", csvFile.Name())

	//panic(errors.New("error"))
}

func runMediaSearchStream(ctx context.Context) error {
	roots := strings.Split(*target, ",")
	hashTypes := makeHashTypes()

	_hashStorage, err := hash.NewStorage(*source, hashTypes)
	if err != nil {
		return err
	}
	hashStorage = _hashStorage

	msstream := fstream.NewMediaSearchStream(ctx, roots, *cpu).
		OnError(func(err error) {
			fmt.Printf("[!] Error: %v\n", err.Error())
		}).
		OnEach(fnEachFilter).
		OnMatch(func(m *media.Media) {
			for _, match := range m.Match() {
				printMatch(match.Name, m.Name(), m.Path(), match.HashType, match.Distance)
			}

			countMatchCh <- struct{}{}
		})

	for _, h := range hashTypes {
		switch h {
		case hash.SHA1:
			msstream.OnEach(fnEachSHA1)
		case hash.ED2K:
			msstream.OnEach(fnEachED2K)
		case hash.AHash:
			msstream.OnEach(fnEachAHash)
		case hash.DHash:
			msstream.OnEach(fnEachDHash)
		case hash.DHashV:
			msstream.OnEach(fnEachDHashV)
		case hash.PHash:
			msstream.OnEach(fnEachPHash)
		default:
			return errors.New("hash not found")
		}
	}

	msstream.Run()

	close(countFileCh)
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

func fnEachFilter(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	if m.Type() == "image" {
		countFileCh <- struct{}{}
		return fstream.Next, nil
	}
	return fstream.Skip, nil
}

func fnEachSHA1(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	h, err := m.SHA1()
	if err != nil {
		return fstream.Skip, err
	}
	if src := hashStorage.FindByHash(hash.SHA1, h); src != "-1" {
		m.AddMatch(src, hash.SHA1.String(), 0)
		return fstream.Match, nil
	}
	return fstream.Next, nil
}

func fnEachED2K(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	h, err := m.ED2K()
	if err != nil {
		return fstream.Skip, err
	}
	if src := hashStorage.FindByHash(hash.ED2K, h); src != "-1" {
		m.AddMatch(src, hash.ED2K.String(), 0)
		return fstream.Match, nil
	}
	return fstream.Next, nil
}

func fnEachAHash(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	h, err := m.AHash()
	if err != nil {
		return fstream.Skip, err
	}
	if dist, src := hashStorage.FindByPerceptualHash(hash.AHash, h[0], *distance); dist != -1 {
		m.AddMatch(src, hash.AHash.String(), dist)
		return fstream.Match, nil
	}
	return fstream.Next, nil
}

func fnEachDHash(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	h, err := m.DHash()
	if err != nil {
		return fstream.Skip, err
	}
	if dist, src := hashStorage.FindByPerceptualHash(hash.DHash, h[0], *distance); dist != -1 {
		m.AddMatch(src, hash.DHash.String(), dist)
		return fstream.Match, nil
	}
	return fstream.Next, nil
}

func fnEachDHashV(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by DHashV
	h, err := m.DHashV()
	if err != nil {
		return fstream.Skip, err
	}
	if dist, src := hashStorage.FindByPerceptualHash(hash.DHashV, h[0], *distance); dist != -1 {
		m.AddMatch(src, hash.DHashV.String(), dist)
		return fstream.Match, nil
	}
	return fstream.Next, nil
}

func fnEachPHash(_ context.Context, m *media.Media) (fstream.ResultType, error) { // Search by PHash
	h, err := m.PHash()
	if err != nil {
		return fstream.Skip, err
	}
	if dist, src := hashStorage.FindByPerceptualHash(hash.PHash, h[0], *distance); dist != -1 {
		m.AddMatch(src, hash.PHash.String(), dist)
		return fstream.Match, nil
	}
	return fstream.Next, nil
}

func printMatch(sourceName, targetName, targetPath, hashType string, distHamming int) {
	_csv.Write([]string{
		sourceName,
		targetName,
		targetPath,
		hashType,
		strconv.Itoa(distHamming),
	})
}

func printBanner() {
	fmt.Println("###############################################################################")
	fmt.Printf("#%78s\n", "#")
	color.Printf("%-35s <yellow>%s</> %36s\n", "#", "ChaSAM", "#")
	fmt.Printf("#%78s\n", "#")
	fmt.Println("###############################################################################")
	color.Println("[>] Para abortar pressione as teclas <red>ctrl+c</>")
	color.Println("[>] Iniciando a busca...\n")
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
