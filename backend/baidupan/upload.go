package baidupan

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rclone/rclone/backend/baidupan/api"
	"github.com/rclone/rclone/fs"
	xpansdk "open-sdk-go/openxpanapi"
)

// uploadFile uploads a file to Baidu Pan using multipart upload
func (f *Fs) uploadFile(ctx context.Context, in io.Reader, size int64, remote string, options ...fs.OpenOption) (*api.CreateFileResponse, error) {
	filePath := f.makePath(remote)

	if size < 0 {
		return nil, fmt.Errorf("cannot upload file with unknown size")
	}

	// Read the entire file into memory for small files, or use a buffer for large files
	var buf bytes.Buffer
	if size <= int64(f.opt.ChunkSize)*4 {
		// For small files, read everything
		if _, err := io.Copy(&buf, in); err != nil {
			return nil, fmt.Errorf("failed to buffer file: %w", err)
		}
		in = &buf
	} else {
		// For large files, we need to use TeeReader to calculate hashes
		if _, err := io.Copy(&buf, io.TeeReader(in, &buf)); err != nil {
			return nil, fmt.Errorf("failed to buffer file: %w", err)
		}
		in = &buf
	}

	// Calculate content MD5
	contentMD5, err := calculateMD5(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MD5: %w", err)
	}

	// Calculate block MD5s
	blockSize := int64(f.opt.ChunkSize)
	blockMD5s, err := calculateBlockMD5s(bytes.NewReader(buf.Bytes()), size, blockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate block MD5s: %w", err)
	}

	// Step 1: Precreate - check if rapid upload is possible
	precreateResp, err := f.precreate(ctx, filePath, size, contentMD5, blockMD5s)
	if err != nil {
		return nil, fmt.Errorf("precreate failed: %w", err)
	}

	// If rapid upload succeeded (return_type == 2), we're done
	if precreateResp.ReturnType == 2 {
		fs.Debugf(f, "Rapid upload succeeded for %s", remote)
		// Get file info by listing
		return f.createFileFinish(ctx, filePath, size, precreateResp.UploadID, blockMD5s)
	}

	// Step 2: Upload blocks that need uploading
	if len(precreateResp.BlockList) > 0 {
		err = f.uploadBlocks(ctx, bytes.NewReader(buf.Bytes()), size, filePath, precreateResp.UploadID, precreateResp.BlockList, blockMD5s)
		if err != nil {
			return nil, fmt.Errorf("upload blocks failed: %w", err)
		}
	}

	// Step 3: Create file
	return f.createFileFinish(ctx, filePath, size, precreateResp.UploadID, blockMD5s)
}

// calculateMD5 calculates the MD5 hash of a reader
func calculateMD5(in io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, in); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// calculateBlockMD5s calculates MD5 hashes for each block
func calculateBlockMD5s(in io.Reader, size int64, blockSize int64) ([]string, error) {
	var blockMD5s []string
	remaining := size

	for remaining > 0 {
		blockLen := blockSize
		if remaining < blockSize {
			blockLen = remaining
		}

		limitReader := io.LimitReader(in, blockLen)
		blockHash := md5.New()
		if _, err := io.Copy(blockHash, limitReader); err != nil {
			return nil, err
		}

		blockMD5s = append(blockMD5s, hex.EncodeToString(blockHash.Sum(nil)))
		remaining -= blockLen
	}

	return blockMD5s, nil
}

// precreate calls the precreate API to prepare for upload
func (f *Fs) precreate(ctx context.Context, path string, size int64, contentMD5 string, blockMD5s []string) (*api.PrecreateResponse, error) {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	blockListJSON, err := json.Marshal(blockMD5s)
	if err != nil {
		return nil, err
	}

	resp, err := callTyped(f, ctx, func() (xpansdk.Fileprecreateresponse, *http.Response, error) {
		return f.sdk.FileuploadApi.Xpanfileprecreate(ctx).
			AccessToken(accessToken).
			Path(path).
			Isdir(0).
			Size(size).
			Autoinit(1).
			BlockList(string(blockListJSON)).
			Rtype(3).
			Execute()
	})
	if err != nil {
		return nil, err
	}
	if resp.GetErrno() != 0 {
		return nil, &api.Error{
			ErrorCode: int(resp.GetErrno()),
			ErrorMsg:  fmt.Sprintf("precreate error code %d", resp.GetErrno()),
		}
	}

	converted := &api.PrecreateResponse{
		UploadID:   resp.GetUploadid(),
		ReturnType: int(resp.GetReturnType()),
	}
	if list := resp.GetBlockList(); len(list) > 0 {
		converted.BlockList = make([]int, len(list))
		for i, v := range list {
			idx, parseErr := strconv.Atoi(v)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid block index %q: %w", v, parseErr)
			}
			converted.BlockList[i] = idx
		}
	}

	return converted, nil
}

// uploadBlocks uploads the specified blocks
func (f *Fs) uploadBlocks(ctx context.Context, in io.Reader, size int64, path string, uploadID string, blockList []int, blockMD5s []string) error {
	var seeker io.ReadSeeker
	if rs, ok := in.(io.ReadSeeker); ok {
		seeker = rs
	} else {
		data, err := io.ReadAll(in)
		if err != nil {
			return fmt.Errorf("failed to buffer upload data: %w", err)
		}
		seeker = bytes.NewReader(data)
	}

	blockSize := int64(f.opt.ChunkSize)

	for _, blockNum := range blockList {
		if blockNum >= len(blockMD5s) {
			return fmt.Errorf("block number %d out of range", blockNum)
		}

		offset := int64(blockNum) * blockSize
		if _, err := seeker.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek to block %d: %w", blockNum, err)
		}

		// Read the block
		blockLen := blockSize
		if offset+blockSize > size {
			blockLen = size - offset
		}

		blockData := make([]byte, blockLen)
		if _, err := io.ReadFull(seeker, blockData); err != nil {
			return fmt.Errorf("failed to read block %d: %w", blockNum, err)
		}

		// Upload the block
		err := f.uploadBlock(ctx, path, blockData, uploadID, blockNum)
		if err != nil {
			return fmt.Errorf("failed to upload block %d: %w", blockNum, err)
		}

		fs.Debugf(f, "Uploaded block %d/%d", blockNum+1, len(blockMD5s))
	}

	return nil
}

// uploadBlock uploads a single block
func (f *Fs) uploadBlock(ctx context.Context, path string, data []byte, uploadID string, blockNum int) error {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return err
	}

	_, err = callTyped(f, ctx, func() (string, *http.Response, error) {
		return f.sdk.FileuploadApi.Pcssuperfile2(ctx).
			AccessToken(accessToken).
			Partseq(strconv.Itoa(blockNum)).
			Path(path).
			Uploadid(uploadID).
			Type_("tmpfile").
			FileReader(bytes.NewReader(data), fmt.Sprintf("chunk-%d", blockNum)).
			Execute()
	})
	return err
}

// createFileFinish calls the create API to finish the upload
func (f *Fs) createFileFinish(ctx context.Context, path string, size int64, uploadID string, blockMD5s []string) (*api.CreateFileResponse, error) {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	blockListJSON, err := json.Marshal(blockMD5s)
	if err != nil {
		return nil, err
	}

	resp, err := callTyped(f, ctx, func() (xpansdk.Filecreateresponse, *http.Response, error) {
		return f.sdk.FileuploadApi.Xpanfilecreate(ctx).
			AccessToken(accessToken).
			Path(path).
			Isdir(0).
			Size(size).
			Uploadid(uploadID).
			BlockList(string(blockListJSON)).
			Rtype(3).
			Execute()
	})
	if err != nil {
		return nil, err
	}
	if resp.GetErrno() != 0 {
		return nil, &api.Error{
			ErrorCode: int(resp.GetErrno()),
			ErrorMsg:  fmt.Sprintf("create error code %d", resp.GetErrno()),
		}
	}

	result := &api.CreateFileResponse{
		FsID:           resp.GetFsId(),
		Path:           resp.GetPath(),
		ServerFilename: resp.GetServerFilename(),
		Category:       int(resp.GetCategory()),
		Size:           resp.GetSize(),
		MD5:            resp.GetMd5(),
		ServerMtime:    resp.GetServerMtime(),
		ServerCtime:    resp.GetServerCtime(),
		LocalMtime:     resp.GetLocalMtime(),
		LocalCtime:     resp.GetLocalCtime(),
		Isdir:          int(resp.GetIsdir()),
	}

	return result, nil
}
