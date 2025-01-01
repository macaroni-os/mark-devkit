/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FetcherS3 struct {
	*FetcherCommon

	MinioClient *minio.Client
	Bucket      string
	Prefix      string
}

func NewFetcherS3(c *specs.MarkDevkitConfig, opts map[string]string) (*FetcherS3, error) {
	if _, ok := opts["minio-bucket"]; !ok {
		return nil, errors.New("Minio bucket is mandatory")
	}

	if _, ok := opts["minio-endpoint"]; !ok {
		return nil, errors.New("Minio endpoint is mandatory")
	}

	if _, ok := opts["minio-keyid"]; !ok {
		return nil, errors.New("Minio key ID is mandatory")
	}

	if _, ok := opts["minio-secret"]; !ok {
		return nil, errors.New("Minio secret Access key is mandatory")
	}

	ans := &FetcherS3{
		FetcherCommon: NewFetcherCommon(c),
		Bucket:        opts["minio-bucket"],
		Prefix:        opts["minio-prefix"],
	}

	minioRegion := ""
	minioSsl := true
	if _, ok := opts["minio-region"]; ok {
		minioRegion = opts["minio-region"]
	}

	if _, ok := opts["minio-ssl"]; ok {
		if opts["minio-ssl"] == "false" {
			minioSsl = false
		}
	}

	var mClient *minio.Client
	var err error

	mOpts := &minio.Options{
		Creds: credentials.NewStaticV4(
			opts["minio-keyid"],
			opts["minio-secret"],
			"",
		),
		Secure: minioSsl,
	}
	if minioRegion != "" {
		mOpts.Region = minioRegion
	}

	mClient, err = minio.New(
		opts["minio-endpoint"],
		mOpts,
	)
	if err != nil {
		return nil, errors.New("Error on create minio client: " + err.Error())
	}

	ans.MinioClient = mClient

	// Check if the bucket exists
	found, err := ans.MinioClient.BucketExists(context.Background(), ans.Bucket)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("Error on check if the bucket %s: %s", ans.Bucket, err.Error()))
	}

	if !found {
		return nil, errors.New(fmt.Sprintf("Bucket %s not found", ans.Bucket))
	}

	return ans, nil
}

func (f *FetcherS3) Sync(specfile string, opts *FetchOpts) error {
	// Load MergeKit specs
	mkit := specs.NewDistfilesSpec()

	if opts.CleanWorkingDir {
		defer os.RemoveAll(f.GetReposcanDir())
		defer os.RemoveAll(f.GetTargetDir())
	}

	err := mkit.LoadFile(specfile)
	if err != nil {
		return err
	}

	err = f.PrepareSourcesKits(mkit, opts)
	if err != nil {
		return err
	}

	err = f.syncAtoms(mkit, opts)
	if err != nil {
		return err
	}

	return nil
}

func (f *FetcherS3) syncAtoms(mkit *specs.DistfilesSpec, opts *FetchOpts) error {

	// Prepare download directory
	err := helpers.EnsureDirWithoutIds(f.GetDownloadDir(), 0755)
	if err != nil {
		return err
	}

	// Retrieve list of existing objects
	mapFilesObjects, err := f.getS3Files(mkit, opts)
	if err != nil {
		return err
	}

	for catpkg, atoms := range f.Resolver.Map {

		f.Logger.Debug(fmt.Sprintf(":factory:[%s] Analyzing ...", catpkg))

		for idx := range atoms {
			f.Logger.Debug(fmt.Sprintf(":factory:[%s] Analyzing ...", atoms[idx].Atom))

			f.Stats.IncrementElab()

			if len(atoms[idx].Files) > 0 {
				err := f.syncAtom(mkit, opts, &atoms[idx], mapFilesObjects)
				if err != nil {
					f.AddAtomInError(&atoms[idx], err)
				}
			} else {
				f.Logger.Debug(fmt.Sprintf(":smiling_face_with_sunglasses:[%s] Nothing to do.", atoms[idx].Atom))
			}
		}
	}

	return nil
}

func (f *FetcherS3) GetFileFromObjectStorage(file, to string) error {
	fd, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("error on create file %s: %s",
			to, err.Error())
	}
	defer fd.Close()

	object, err := f.MinioClient.GetObject(
		context.Background(), f.Bucket, file, minio.GetObjectOptions{},
	)
	if err != nil {
		return err
	}

	if _, err = io.Copy(fd, object); err != nil {
		return err
	}

	return nil
}

func (f *FetcherS3) RemoveFileFromObjectStorage(file string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	return f.MinioClient.RemoveObject(context.Background(),
		f.Bucket, file, opts)
}

func (f *FetcherS3) UploadFile2ObjectStorage(atom *specs.RepoScanAtom, file, s3path string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	fileStat, err := fd.Stat()
	if err != nil {
		return err
	}

	uploadInfo, err := f.MinioClient.PutObject(context.Background(),
		f.Bucket, s3path, fd, fileStat.Size(),
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
	if err != nil {
		return err
	}

	f.Logger.InfoC(fmt.Sprintf(
		":telescope:[%s] Uploaded file %s of size %d.",
		atom.Atom, filepath.Base(file), uploadInfo.Size))

	return nil
}

func (f *FetcherS3) syncAtom(mkit *specs.DistfilesSpec, opts *FetchOpts,
	atom *specs.RepoScanAtom, mapFilesObjects *map[string]*minio.ObjectInfo) error {

	toUpload := false
	toDownload := false
	filesMap := make(map[string]string, 0)
	files2Remove := make(map[string]*specs.RepoScanFile, 0)
	filesOk := make(map[string]*specs.RepoScanFile, 0)
	atomSize := int64(0)

	mapS3files := *mapFilesObjects

	// Check if all files are availables on S3 Object store
	for _, file := range atom.Files {

		s3objectPath := filepath.Join(f.Prefix, file.Name)
		downloadedFilePath := filepath.Join(f.GetDownloadDir(), file.Name)

		// Check if the file is already present
		if _, present := filesMap[file.Name]; present {
			// File already processed
			continue
		} else {
			size, _ := strconv.ParseInt(file.Size, 10, 64)
			atomSize += size
		}

		filesMap[file.Name] = file.Size

		// Check if the file is available on Object store.
		if oinfo, present := mapS3files[s3objectPath]; present {

			size, _ := strconv.ParseInt(file.Size, 10, 64)
			// Check if size is equal
			if size != oinfo.Size {
				f.Logger.Warning(fmt.Sprintf(
					":warning:[%s] File %s with size %d instead of %d.",
					atom.Atom, file.Name, oinfo.Size, file.Size,
				))
				toUpload = true
				files2Remove[s3objectPath] = &file
				continue
			}

			// Check if md5 is equal
			md5Hash, withMd5 := file.Hashes["md5"]

			if withMd5 && md5Hash != oinfo.ETag {
				f.Logger.Warning(fmt.Sprintf(
					":warning:[%s] File %s with md5 %s instead of %s.",
					atom.Atom, file.Name, oinfo.ETag, md5Hash,
				))
				toUpload = true
				files2Remove[s3objectPath] = &file
				continue
			} else if !withMd5 {

				// Download file from S3 Object Storage.
				err := f.GetFileFromObjectStorage(s3objectPath, downloadedFilePath)
				if err != nil {
					return err
				}

				md5Hash, err := helpers.GetFileMd5(downloadedFilePath)
				if err != nil {
					return err
				}

				if md5Hash != oinfo.ETag {
					f.Logger.Warning(fmt.Sprintf(
						":warning:[%s] File %s downloaded with md5 %s instead of %s.",
						atom.Atom, file.Name, oinfo.ETag, md5Hash,
					))
					toUpload = true
					files2Remove[s3objectPath] = &file
					continue
				}

			} else {
				// POST: else md5 is correct
				filesOk[s3objectPath] = &file
			}

		} else {
			// POST: file s3object not present.
			toUpload = true
		}

		if toUpload {
			if !utils.Exists(downloadedFilePath) {
				toDownload = true
				// Skip break to get all files size
				// break
			}
		}
	}

	if atomSize > 0 {
		f.Stats.IncrementSize(atomSize)
	}

	if toDownload {
		f.Logger.InfoC(
			fmt.Sprintf(":factory:[%s] Downloading files...", atom.Atom))

		err := f.DownloadAtomsFiles(mkit, atom)
		if err != nil {
			f.Stats.IncrementErrors()
			return err
		}

		f.Stats.IncrementAtoms()

		f.Logger.InfoC(
			fmt.Sprintf(":medal: [%s] Files downloaded.", atom.Atom))

		// Add all files ok to remove list. The download phase is
		// done for all files of the atom.
		for s3objectPath, file := range filesOk {
			files2Remove[s3objectPath] = file
		}
	}

	if toUpload {
		if len(files2Remove) > 0 {
			for s3objectPath, file := range files2Remove {
				err := f.RemoveFileFromObjectStorage(s3objectPath)
				if err != nil {
					return err
				}
				f.Logger.Debug(fmt.Sprintf(
					":knife:[%s] Removed file %s (%s) from S3 Object Storage.",
					atom.Atom, s3objectPath, file.Name,
				))
			}
		}

		filesMap = make(map[string]string, 0)

		for _, file := range atom.Files {
			s3objectPath := filepath.Join(f.Prefix, file.Name)
			downloadedFilePath := filepath.Join(f.GetDownloadDir(), file.Name)
			// Check if the file is already present
			if _, present := filesMap[file.Name]; present {
				// File already processed
				continue
			}
			filesMap[file.Name] = file.Size

			err := f.UploadFile2ObjectStorage(atom, downloadedFilePath,
				s3objectPath,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FetcherS3) getS3Files(mkit *specs.DistfilesSpec,
	opts *FetchOpts) (*map[string]*minio.ObjectInfo, error) {

	ans := make(map[string]*minio.ObjectInfo, 0)

	listOpts := minio.ListObjectsOptions{
		Recursive:    true,
		Prefix:       f.Prefix,
		WithMetadata: true,
	}

	ctx := context.Background()

	// List all objects from a bucket-name with a matching prefix.
	for object := range f.MinioClient.ListObjects(ctx, f.Bucket, listOpts) {
		if object.Err != nil {
			return &ans, fmt.Errorf("error on retrieve list of objects: %s",
				object.Err.Error())
		}

		ans[object.Key] = &object
	}

	// NOTE: GetObjectAttributes method is not supported by CDN77 object storage.

	return &ans, nil
}
