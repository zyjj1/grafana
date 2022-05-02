<<<<<<< HEAD
import React, { Dispatch, SetStateAction } from 'react';
import { FileDropzone, useTheme2, Button, DropzoneFile } from '@grafana/ui';
import { getBackendSrv, config } from '@grafana/runtime';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { MediaType } from '../types';
interface Props {
  setNewValue: Dispatch<SetStateAction<string>>;
  setFormData: Dispatch<SetStateAction<FormData>>;
  mediaType: MediaType;
  setUpload: Dispatch<SetStateAction<boolean>>;
  getRequest: (formData: FormData) => Promise<UploadResponse>;
}
interface UploadResponse {
  err: boolean;
  path: string;
}
export function FileDropzoneCustomChildren({ secondaryText = 'Drag and drop here or browse' }) {
  const theme = useTheme2();
  const styles = getStyles(theme);
=======
import { css } from '@emotion/css';
import React, { Dispatch, SetStateAction, useState } from 'react';
import SVG from 'react-inlinesvg';

import { GrafanaTheme2 } from '@grafana/data';
import { FileDropzone, useStyles2, Button, DropzoneFile, Field } from '@grafana/ui';

import { MediaType } from '../types';
interface Props {
  setFormData: Dispatch<SetStateAction<FormData>>;
  mediaType: MediaType;
  setUpload: Dispatch<SetStateAction<boolean>>;
  newValue: string;
  error: ErrorResponse;
}
interface ErrorResponse {
  message: string;
}
export function FileDropzoneCustomChildren({ secondaryText = 'Drag and drop here or browse' }) {
  const styles = useStyles2(getStyles);
>>>>>>> main

  return (
    <div className={styles.iconWrapper}>
      <small className={styles.small}>{secondaryText}</small>
<<<<<<< HEAD
      <Button icon="upload">Upload</Button>
    </div>
  );
}
export const FileUploader = ({ mediaType, setNewValue, setFormData, setUpload, getRequest }: Props) => {
=======
      <Button type="button" icon="upload">
        Upload
      </Button>
    </div>
  );
}
export const FileUploader = ({ mediaType, setFormData, setUpload, error }: Props) => {
  const styles = useStyles2(getStyles);
  const [dropped, setDropped] = useState<boolean>(false);
  const [file, setFile] = useState<string>('');

  const Preview = () => (
    <Field label="Preview">
      <div className={styles.iconPreview}>
        {mediaType === MediaType.Icon && <SVG src={file} className={styles.img} />}
        {mediaType === MediaType.Image && <img src={file} className={styles.img} />}
      </div>
    </Field>
  );

>>>>>>> main
  const onFileRemove = (file: DropzoneFile) => {
    fetch(`/api/storage/delete/upload/${file.file.name}`, {
      method: 'DELETE',
    }).catch((error) => console.error('cannot delete file', error));
  };
<<<<<<< HEAD
=======

>>>>>>> main
  const acceptableFiles =
    mediaType === 'icon' ? 'image/svg+xml' : 'image/jpeg,image/png,image/gif,image/png, image/webp';
  return (
    <FileDropzone
      readAs="readAsBinaryString"
      onFileRemove={onFileRemove}
      options={{
        accept: acceptableFiles,
        multiple: false,
        onDrop: (acceptedFiles: File[]) => {
          let formData = new FormData();
          formData.append('file', acceptedFiles[0]);
<<<<<<< HEAD
          getRequest(formData).then((data) => {
            if (!data.err) {
              getBackendSrv()
                .get(`api/storage/read/${data.path}`)
                .then(() => {
                  setNewValue(`${config.appUrl}api/storage/read/${data.path}`);
                });
            }
          });
=======
          setFile(URL.createObjectURL(acceptedFiles[0]));
          setDropped(true);
>>>>>>> main
          setFormData(formData);
          setUpload(true);
        },
      }}
    >
<<<<<<< HEAD
      <FileDropzoneCustomChildren />
=======
      {error.message !== '' && dropped ? (
        <p>{error.message}</p>
      ) : dropped ? (
        <Preview />
      ) : (
        <FileDropzoneCustomChildren />
      )}
>>>>>>> main
    </FileDropzone>
  );
};

function getStyles(theme: GrafanaTheme2, isDragActive?: boolean) {
  return {
    container: css`
      display: flex;
      flex-direction: column;
      width: 100%;
    `,
    dropzone: css`
      display: flex;
      flex: 1;
      flex-direction: column;
      align-items: center;
      padding: ${theme.spacing(6)};
      border-radius: 2px;
      border: 2px dashed ${theme.colors.border.medium};
      background-color: ${isDragActive ? theme.colors.background.secondary : theme.colors.background.primary};
      cursor: pointer;
    `,
    iconWrapper: css`
      display: flex;
      flex-direction: column;
      align-items: center;
    `,
    acceptMargin: css`
      margin: ${theme.spacing(2, 0, 1)};
    `,
    small: css`
      color: ${theme.colors.text.secondary};
      margin-bottom: ${theme.spacing(2)};
    `,
<<<<<<< HEAD
=======
    iconContainer: css`
      display: flex;
      flex-direction: column;
      width: 80%;
      align-items: center;
      align-self: center;
    `,
    iconPreview: css`
      width: 238px;
      height: 198px;
      border: 1px solid ${theme.colors.border.medium};
      display: flex;
      align-items: center;
      justify-content: center;
    `,
    img: css`
      width: 147px;
      height: 147px;
      fill: ${theme.colors.text.primary};
    `,
>>>>>>> main
  };
}
